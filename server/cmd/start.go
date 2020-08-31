/*
Copyright © 2020 Valeriy Molchanov <valeriy.molchanov.77@gmail.com>

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

// Package cmd defines commands which server can do.
package cmd

import (
	"fmt"
	"net"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
	"github.com/movaua/rock-paper-scissors/pkg/server"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts Rock Paper Scissoers game server",
	RunE:  startServer,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVarP(&port, "port", "p", 9090, "game server port")
	startCmd.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 10, "player answer timeout, seconds")
}

var (
	port           int
	timeoutSeconds int
)

func startServer(cmd *cobra.Command, args []string) error {
	// cmd.SilenceUsage = true

	fmt.Printf("Player answer timeout is %d seconds\n", timeoutSeconds)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("starting game server at %s\n", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot listen %s: %w", addr, err)
	}

	fmt.Printf("listening  %q\n", addr)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	roundTimeout := time.Duration(timeoutSeconds) * time.Second
	gameServer := server.NewGame(cmd.Context(), roundTimeout)

	pb.RegisterGamerServer(grpcServer, gameServer)

	return grpcServer.Serve(lis)
}
