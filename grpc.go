package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"xquant/pkg/utils"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
)

func runGrpc() {
	if !env.GrpcIsOpen {
		return
	}

	port := utils.GetEnvWithDefault("PORT_OF_GRPC", "8890")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}

	svr := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(middleware.RecoverHandler)),
	), grpc.ChainStreamInterceptor(
		recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(middleware.RecoverHandler)),
	))
	api_model.RegisterApiServiceServer(svr, &api_handler.GrpcServiceImpl{})
	scheduler_model.RegisterMaasNodeServiceServer(svr, &scheduler_handler.GrpcServiceImpl{})
	reflection.Register(svr)

	sig, wg := make(chan os.Signal, 1), sync.WaitGroup{}
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sig
		log.Infof("GRPC: receive signal %v, attempting graceful shutdown", s)
		svr.GracefulStop()
	}()

	log.Infof("start grpc server at %s", lis.Addr())
	if err := svr.Serve(lis); err != nil {
		panic(err)
	}
}
