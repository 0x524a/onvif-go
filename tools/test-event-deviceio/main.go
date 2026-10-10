// Package main tests Event and Device IO services against a real camera.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	onvif "github.com/0x524a/onvif-go"
)

const notAvailable = "N/A"

func main() {
	// Command line flags.
	cameraIP := flag.String("ip", "192.168.1.201", "Camera IP address")
	username := flag.String("user", "service", "Camera username")
	password := flag.String("pass", "Service.1234", "Camera password")
	flag.Parse()

	endpoint := fmt.Sprintf("http://%s/onvif/device_service", *cameraIP)

	fmt.Printf("Testing Event and Device IO services on camera: %s\n", *cameraIP)
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Username: %s\n\n", *username)

	// Create client.
	client, err := onvif.NewClient(endpoint,
		onvif.WithCredentials(*username, *password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Test device information first to verify connectivity.
	fmt.Println("=== Testing Device Connectivity ===")
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		fmt.Printf("Failed to get device information: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Device: %s %s\n", info.Manufacturer, info.Model)
	fmt.Printf("Firmware: %s\n", info.FirmwareVersion)
	fmt.Printf("Serial: %s\n\n", info.SerialNumber)

	// Test Event Service.
	testEventService(ctx, client)

	// Test Device IO Service.
	testDeviceIOService(ctx, client)

	fmt.Println("\n=== All Tests Completed ===")
}

func testEventService(ctx context.Context, client *onvif.Client) {
	fmt.Println("=== Testing Event Service ===")

	// 1. Get Event Service Capabilities.
	fmt.Println("\n1. GetEventServiceCapabilities")
	caps, err := client.GetEventServiceCapabilities(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   WSSubscriptionPolicySupport: %v\n", caps.WSSubscriptionPolicySupport)
		fmt.Printf("   MaxPullPoints: %d\n", caps.MaxPullPoints)
		fmt.Printf("   PersistentNotificationStorage: %v\n", caps.PersistentNotificationStorage)
		fmt.Printf("   EventBrokerProtocols: %v\n", caps.EventBrokerProtocols)
		fmt.Printf("   MaxEventBrokers: %d\n", caps.MaxEventBrokers)
	}

	// 2. Get Event Properties.
	fmt.Println("\n2. GetEventProperties")
	props, err := client.GetEventProperties(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   FixedTopicSet: %v\n", props.FixedTopicSet)
		fmt.Printf("   TopicNamespaceLocations: %d\n", len(props.TopicNamespaceLocation))
		fmt.Printf("   TopicExpressionDialects: %d\n", len(props.TopicExpressionDialects))
	}

	// 3. Create Pull Point Subscription.
	fmt.Println("\n3. CreatePullPointSubscription")
	termTime := 60 * time.Second
	sub, err := client.CreatePullPointSubscription(ctx, "", &termTime, "")
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   SubscriptionReference: %s\n", sub.SubscriptionReference)
		fmt.Printf("   CurrentTime: %v\n", sub.CurrentTime)
		fmt.Printf("   TerminationTime: %v\n", sub.TerminationTime)

		// 4. Pull Messages.
		if sub.SubscriptionReference != "" {
			fmt.Println("\n4. PullMessages")
			messages, err := client.PullMessages(ctx, sub.SubscriptionReference, 5*time.Second, 10)
			if err != nil {
				fmt.Printf("   ERROR: %v\n", err)
			} else {
				fmt.Printf("   Received %d messages\n", len(messages))
				for i, msg := range messages {
					if i >= 3 {
						fmt.Printf("   ... and %d more\n", len(messages)-3)
						break
					}

					fmt.Printf("   Message %d: Topic=%s, Operation=%s\n",
						i+1, msg.Topic, msg.Message.PropertyOperation)
				}
			}

			// 5. Renew Subscription.
			fmt.Println("\n5. RenewSubscription")
			curTime, newTermTime, err := client.RenewSubscription(ctx, sub.SubscriptionReference, 120*time.Second)
			if err != nil {
				fmt.Printf("   ERROR: %v\n", err)
			} else {
				fmt.Printf("   CurrentTime: %v\n", curTime)
				fmt.Printf("   NewTerminationTime: %v\n", newTermTime)
			}

			// 6. Unsubscribe.
			fmt.Println("\n6. Unsubscribe")
			err = client.Unsubscribe(ctx, sub.SubscriptionReference)
			if err != nil {
				fmt.Printf("   ERROR: %v\n", err)
			} else {
				fmt.Println("   Successfully unsubscribed")
			}
		}
	}

	// 7. Get Event Brokers (optional, may not be supported).
	fmt.Println("\n7. GetEventBrokers")
	brokers, err := client.GetEventBrokers(ctx)
	if err != nil {
		fmt.Printf("   ERROR (may not be supported): %v\n", err)
	} else {
		fmt.Printf("   Found %d event brokers\n", len(brokers))
		for i, broker := range brokers {
			fmt.Printf("   Broker %d: %s (Status: %s)\n", i+1, broker.Address, broker.Status)
		}
	}
}

func testDeviceIOService(ctx context.Context, client *onvif.Client) {
	fmt.Println("\n=== Testing Device IO Service ===")

	// 1. Get Device IO Service Capabilities.
	fmt.Println("\n1. GetDeviceIOServiceCapabilities")
	caps, err := client.GetDeviceIOServiceCapabilities(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   VideoSources: %d\n", caps.VideoSources)
		fmt.Printf("   VideoOutputs: %d\n", caps.VideoOutputs)
		fmt.Printf("   AudioSources: %d\n", caps.AudioSources)
		fmt.Printf("   AudioOutputs: %d\n", caps.AudioOutputs)
		fmt.Printf("   RelayOutputs: %d\n", caps.RelayOutputs)
		fmt.Printf("   DigitalInputs: %d\n", caps.DigitalInputs)
		fmt.Printf("   SerialPorts: %d\n", caps.SerialPorts)
	}

	// 2. Get Digital Inputs.
	fmt.Println("\n2. GetDigitalInputs")
	inputs, err := client.GetDigitalInputs(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Found %d digital inputs\n", len(inputs))
		for i, input := range inputs {
			fmt.Printf("   Input %d: Token=%s, IdleState=%s\n", i+1, input.Token, input.IdleState)
		}
	}

	// 3. Get Video Outputs.
	fmt.Println("\n3. GetVideoOutputs")
	outputs, err := client.GetVideoOutputs(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Found %d video outputs\n", len(outputs))
		for i, output := range outputs {
			res := notAvailable
			if output.Resolution != nil {
				res = fmt.Sprintf("%dx%d", output.Resolution.Width, output.Resolution.Height)
			}

			fmt.Printf("   Output %d: Token=%s, Resolution=%s, RefreshRate=%.1f\n",
				i+1, output.Token, res, output.RefreshRate)
		}
	}

	// 4. Get Serial Ports.
	fmt.Println("\n4. GetSerialPorts")
	ports, err := client.GetSerialPorts(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Found %d serial ports\n", len(ports))
		for i, port := range ports {
			fmt.Printf("   Port %d: Token=%s, Type=%s\n", i+1, port.Token, port.Type)
		}
	}

	// 5. Get Relay Outputs (using existing method).
	fmt.Println("\n5. GetRelayOutputs")
	relays, err := client.GetRelayOutputs(ctx)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Found %d relay outputs\n", len(relays))
		for i, relay := range relays {
			mode := notAvailable
			idleState := notAvailable
			if relay.Properties.Mode != "" {
				mode = string(relay.Properties.Mode)
			}

			if relay.Properties.IdleState != "" {
				idleState = string(relay.Properties.IdleState)
			}

			fmt.Printf("   Relay %d: Token=%s, Mode=%s, IdleState=%s\n",
				i+1, relay.Token, mode, idleState)
		}
	}
}
