package controllers

import (
	"context"
	"net/http"
	"shelly/config"
	"shelly/middleware"
	"shelly/models"
	"time"

	"io"
	"log"
	"strings"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func CreateShell(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDProp)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User ID not found"})
		return
	}

	var existingShell models.Shell
	err := config.ShellCollection.FindOne(c, bson.M{"user_id": userID.(string)}).Decode(&existingShell)
	if err == nil {
		c.JSON(http.StatusAccepted, gin.H{"success": true, "message": "User already has a shell"})
		return
	}
	shellUUID := uuid.New().String()

	shell := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shellUUID,
			Namespace: "default",
			Labels: map[string]string{
				"app":     "ubuntu-shell",
				"user-id": userID.(string),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "ubuntu",
					Image: "ubuntu:latest",
					Command: []string{
						"/bin/bash",
						"-c",
						"while true; do sleep 3600; done",
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := config.Clientset.CoreV1().Pods("default").Create(ctx, shell, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	shellDoc := models.Shell{
		UserID: userID.(string),
		UUID:   shellUUID,
	}
	_, err = config.ShellCollection.InsertOne(c, shellDoc)
	if err != nil {
		config.Clientset.CoreV1().Pods("default").Delete(ctx, shellUUID, metav1.DeleteOptions{})
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to store shell information"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "shell created successfully",
		"shell": gin.H{
			"uuid":      result.Name,
			"namespace": result.Namespace,
			"status":    result.Status.Phase,
		},
	})
}

func DeleteShell(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDProp)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User ID not found"})
		return
	}

	var shell models.Shell
	err := config.ShellCollection.FindOne(c, bson.M{"user_id": userID.(string)}).Decode(&shell)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "No shell found for this user"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = config.Clientset.CoreV1().Pods("default").Delete(ctx, shell.UUID, metav1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	_, err = config.ShellCollection.DeleteOne(c, bson.M{"user_id": userID.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete shell information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "shell deleted successfully",
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleShellWebSocket(c *gin.Context) {

	if !websocket.IsWebSocketUpgrade(c.Request) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a WebSocket request"})
		return
	}

	userID, exists := c.Get(middleware.UserIDProp)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var shell models.Shell
	err := config.ShellCollection.FindOne(c, bson.M{"user_id": userID.(string)}).Decode(&shell)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No shell found for this user"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer func() {
		conn.Close()
	}()

	ctx := context.Background()

	req := config.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(shell.UUID).
		Namespace("default").
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"/bin/bash"},
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)

	kubeConfig, err := config.GetKubernetesConfig()
	if err != nil {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to get Kubernetes config"))
		return
	}

	executor, err := remotecommand.NewSPDYExecutor(kubeConfig, "POST", req.URL())
	if err != nil {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to create executor"))
		return
	}

	log.Println("Starting exec session")
	stream := &streamHandler{
		conn: conn,
	}

	stdinReader, stdinWriter := io.Pipe()
	defer stdinReader.Close()
	defer stdinWriter.Close()

	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading WebSocket message: %v", err)
				return
			}
			log.Printf("Received message type: %d, content: %s", messageType, string(message))

			if !strings.HasSuffix(string(message), "\n") {
				message = append(message, '\n')
			}

			if _, err := stdinWriter.Write(message); err != nil {
				log.Printf("Error writing to stdin pipe: %v", err)
				return
			}
			log.Printf("Successfully wrote message to stdin pipe")
		}
	}()

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  stdinReader,
		Stdout: stream,
		Stderr: stream,
		Tty:    true,
	})
	if err != nil {
		log.Printf("Failed to start stream: %v", err)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to start stream"))
		return
	}
	log.Println("Exec session started successfully")

	// Keep the main goroutine alive
	select {}
}

type streamHandler struct {
	conn *websocket.Conn
}

func (h *streamHandler) Read(p []byte) (int, error) {
	// This method is called when the shell needs input
	// We don't need to implement this as we're handling input in the goroutine
	return 0, nil
}

func (h *streamHandler) Write(p []byte) (int, error) {
	log.Printf("Received output from shell: %s", string(p))
	// Send the actual command output to the WebSocket
	err := h.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Printf("Error writing to WebSocket: %v", err)
		return 0, err
	}
	log.Printf("Successfully wrote output to WebSocket")
	return len(p), nil
}
