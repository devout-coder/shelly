package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"shelly/config"
	"shelly/middleware"
	"shelly/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *kubernetes.Clientset

func InitKubernetesClient() error {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		log.Printf("In-cluster config failed: %v", err)

		// If in-cluster config fails, try to use kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		log.Printf("Using kubeconfig at: %s", kubeconfig)

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Printf("Failed to build config from kubeconfig: %v", err)
			return err
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create clientset: %v", err)
		return err
	}

	return nil
}

func CreateShell(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDProp)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var existingShell models.Shell
	err := config.ShellCollection.FindOne(c, bson.M{"user_id": userID.(string)}).Decode(&existingShell)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already has a shell"})
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

	result, err := clientset.CoreV1().Pods("default").Create(ctx, shell, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shellDoc := models.Shell{
		UserID: userID.(string),
		UUID:   shellUUID,
	}
	_, err = config.ShellCollection.InsertOne(c, shellDoc)
	if err != nil {
		clientset.CoreV1().Pods("default").Delete(ctx, shellUUID, metav1.DeleteOptions{})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store shell information"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var shell models.Shell
	err := config.ShellCollection.FindOne(c, bson.M{"user_id": userID.(string)}).Decode(&shell)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No shell found for this user"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = clientset.CoreV1().Pods("default").Delete(ctx, shell.UUID, metav1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = config.ShellCollection.DeleteOne(c, bson.M{"user_id": userID.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete shell information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "shell deleted successfully",
	})
}
