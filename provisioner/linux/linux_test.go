package linux

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/metrue/fx/provisioner"
	"github.com/metrue/go-ssh-client"
	sshMocks "github.com/metrue/go-ssh-client/mocks"
)

func TestDriverProvision(t *testing.T) {
	t.Run("SSHConnectError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sshClient := sshMocks.NewMockClienter(ctrl)
		n := &Docker{sshClient: sshClient}
		err := errors.New("could not connect to host")
		sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(false, err).AnyTimes()
		if err := n.Provision(context.Background(), true); err == nil {
			t.Fatalf("should get error when SSH connection not ok")
		}
	})

	t.Run("SSHConnectionNotOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sshClient := sshMocks.NewMockClienter(ctrl)
		n := New(sshClient)
		sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(false, nil).AnyTimes()
		if err := n.Provision(context.Background(), true); err == nil {
			t.Fatalf("should get error when SSH connection not ok")
		}
	})

	t.Run("DockerAndFxAgentOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sshClient := sshMocks.NewMockClienter(ctrl)
		n := New(sshClient)
		sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(true, nil).AnyTimes()
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_fx_agent"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		if err := n.Provision(context.Background(), true); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DockerNotReady", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sshClient := sshMocks.NewMockClienter(ctrl)
		n := New(sshClient)
		sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(true, nil).AnyTimes()
		err := errors.New("docker command not found")
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(err)
		sshClient.EXPECT().RunCommand(scripts["has_docker"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(err)
		sshClient.EXPECT().RunCommand(scripts["install_docker"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["start_dockerd"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_fx_agent"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		if err := n.Provision(context.Background(), true); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FxAgentNotReady", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sshClient := sshMocks.NewMockClienter(ctrl)
		n := New(sshClient)
		sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(true, nil).AnyTimes()
		err := errors.New("fx agent not found")
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_fx_agent"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(err)
		sshClient.EXPECT().RunCommand(scripts["start_fx_agent"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		if err := n.Provision(context.Background(), true); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DockerAndFxAgentNotReady", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sshClient := sshMocks.NewMockClienter(ctrl)
		n := New(sshClient)

		sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(true, nil).AnyTimes()
		err2 := errors.New("fx agent not found")
		err1 := errors.New("docker command not found")
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(err1)
		sshClient.EXPECT().RunCommand(scripts["has_docker"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(err1)
		sshClient.EXPECT().RunCommand(scripts["install_docker"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["start_dockerd"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_fx_agent"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(err2)
		sshClient.EXPECT().RunCommand(scripts["start_fx_agent"].(string), ssh.CommandOptions{Timeout: provisioner.SSHConnectionTimeout}).Return(nil)
		if err := n.Provision(context.Background(), true); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRunCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sshClient := sshMocks.NewMockClienter(ctrl)
	n := &Docker{
		sshClient: sshClient,
	}
	script := "script"
	option := ssh.CommandOptions{
		Timeout: provisioner.SSHConnectionTimeout,
	}
	sshClient.EXPECT().Connectable(provisioner.SSHConnectionTimeout).Return(true, nil)
	sshClient.EXPECT().RunCommand(script, option).Return(nil)
	if err := n.runCmd(script, true, option); err != nil {
		t.Fatal(err)
	}
}
