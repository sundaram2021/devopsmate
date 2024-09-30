package pkg

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/civo/civogo"
)

// SoftwareInstaller interface defines the behavior for installing software
type SoftwareInstaller interface {
	Install(ctx context.Context, instance InstanceDetails) error
}

// InstanceDetails holds information about the instance
type InstanceDetails struct {
	PublicIP string
	SSHKey   string
	Password string
}

// CreateComputeInstance creates the Civo instance and fetches instance details
func CreateComputeInstance(apiKey string, regionCode string, sshKeyPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create channels to communicate instance details and errors
	instanceDetailsCh := make(chan InstanceDetails)
	errCh := make(chan error)

	// Launch goroutine to create and get instance details
	go func() {
		client, err := civogo.NewClient(apiKey, regionCode)
		if err != nil {
			errCh <- fmt.Errorf("Error in creating Civo client: %w", err)
			return
		}

		// Create a new instance configuration
		config, err := client.NewInstanceConfig()
		if err != nil {
			errCh <- fmt.Errorf("Failed to create a new instance config: %w", err)
			return
		}

		// Create the instance
		instance, err := client.CreateInstance(config)
		if err != nil {
			errCh <- fmt.Errorf("Failed to create a new instance: %w", err)
			return
		}

		// Poll for instance to become active and get details
		var publicIP, password string
		for {
			instanceDetails, err := client.GetInstance(instance.ID)
			if err != nil {
				errCh <- fmt.Errorf("Error retrieving instance details: %w", err)
				return
			}

			// Check if the instance is active and retrieve details
			if instanceDetails.Status == "ACTIVE" {
				publicIP = instanceDetails.PublicIP
				password = instanceDetails.InitialPassword // Assuming this is auto-generated

				instanceDetailsCh <- InstanceDetails{
					PublicIP: publicIP,
					SSHKey:   sshKeyPath, // Use the passed SSH key path
					Password: password,
				}
				break
			}

			// Sleep for a few seconds before checking again
			time.Sleep(5 * time.Second)
		}
	}()

	// Wait for either instance details or an error
	select {
	case <-ctx.Done():
		fmt.Println("Operation timed out")
		return
	case err := <-errCh:
		fmt.Printf("Error: %s\n", err)
		return
	case instanceDetails := <-instanceDetailsCh:
		fmt.Printf("Instance created with Public IP: %s\n", instanceDetails.PublicIP)
		fmt.Printf("SSH Key: %s\n", instanceDetails.SSHKey)
		fmt.Printf("Password: %s\n", instanceDetails.Password)

		// Install software using interfaces for modularity
		installers := []SoftwareInstaller{
			&JenkinsInstaller{},
			&SonarQubeInstaller{},
			&BuildPackInstaller{},
			&CivoKubernetesInstaller{},
		}

		for _, installer := range installers {
			err := installer.Install(ctx, instanceDetails)
			if err != nil {
				fmt.Printf("Error installing software: %s\n", err)
				return
			}
		}
	}
}

// JenkinsInstaller struct for Jenkins installation
type JenkinsInstaller struct{}

// Install installs Jenkins on the instance
func (j *JenkinsInstaller) Install(ctx context.Context, instance InstanceDetails) error {
	fmt.Println("Installing Jenkins...")
	cmd := exec.CommandContext(ctx, "ssh", "-v", "-o", "StrictHostKeyChecking=no", "-i", instance.SSHKey, fmt.Sprintf("civo@%s", instance.PublicIP), `
		sudo apt-get update -y
		sudo apt-get install -y openjdk-11-jdk wget
		wget -q -O - https://pkg.jenkins.io/debian/jenkins.io.key | sudo apt-key add -
		echo "deb http://pkg.jenkins.io/debian-stable binary/" | sudo tee /etc/apt/sources.list.d/jenkins.list
		sudo apt-get update -y
		sudo apt-get install -y jenkins
		sudo systemctl start jenkins
	`)
	return cmd.Run()
}

// SonarQubeInstaller struct for SonarQube installation
type SonarQubeInstaller struct{}

// Install installs SonarQube on the instance
func (s *SonarQubeInstaller) Install(ctx context.Context, instance InstanceDetails) error {
	fmt.Println("Installing SonarQube...")
	cmd := exec.CommandContext(ctx, "ssh", "-v", "-o", "StrictHostKeyChecking=no", "-i", instance.SSHKey, fmt.Sprintf("civo@%s", instance.PublicIP), `
		sudo apt-get install -y sonarqube
	`)
	return cmd.Run()
}

// BuildPackInstaller struct for Buildpack installation
type BuildPackInstaller struct{}

// Install installs Buildpack for Dockerfile generation on the instance
func (b *BuildPackInstaller) Install(ctx context.Context, instance InstanceDetails) error {
	fmt.Println("Installing Buildpack...")
	cmd := exec.CommandContext(ctx, "ssh", "-v", "-o", "StrictHostKeyChecking=no", "-i", instance.SSHKey, fmt.Sprintf("civo@%s", instance.PublicIP), `
		curl -s https://buildpacks.io/install.sh | sudo bash
	`)
	return cmd.Run()
}

// CivoKubernetesInstaller struct for Civo Kubernetes Cluster installation
type CivoKubernetesInstaller struct{}

// Install installs Civo CLI and sets up a Kubernetes cluster on the instance
func (c *CivoKubernetesInstaller) Install(ctx context.Context, instance InstanceDetails) error {
	fmt.Println("Installing Civo CLI and Kubernetes Cluster...")
	cmd := exec.CommandContext(ctx, "ssh", "-v", "-o", "StrictHostKeyChecking=no", "-i", instance.SSHKey, fmt.Sprintf("civo@%s", instance.PublicIP), `
		curl -sL https://cli.civo.com/install | sudo bash
		civo kubernetes create my-cluster --size=g3.k3s.medium --nodes=3 --wait
		civo kubernetes config my-cluster --save --local-path ~/.kube/config
		kubectl get nodes
	`)
	return cmd.Run()
}
