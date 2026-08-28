package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"offlinebundle/internal/fetcher"
	"offlinebundle/internal/service"
	"offlinebundle/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: bundle create|inspect")
		return
	}
	switch os.Args[1] {
	case "create":
		if err := runCreate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "inspect":
		if err := runInspect(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(2)
	}
}

func runCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	id := fs.String("id", "job-1", "job identifier")
	root := fs.String("root", "", "root URL")
	output := fs.String("output", "bundle-output", "output directory")
	dbPath := fs.String("db", "bundle.db", "database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *root == "" {
		return fmt.Errorf("-root is required")
	}
	store, err := storage.Open(*dbPath)
	if err != nil {
		return err
	}
	defer store.Close()
	svc := service.New(store, &http.Client{})
	result, err := svc.CreateBundle(*id, *root, *output)
	if err != nil {
		return err
	}
	fmt.Printf("job=%s status=%s incomplete=%t output=%s\n", result.Job.ID, result.Job.Status, result.Manifest.Incomplete, result.Job.OutputDir)
	return nil
}

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	id := fs.String("id", "job-1", "job identifier")
	dbPath := fs.String("db", "bundle.db", "database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	store, err := storage.Open(*dbPath)
	if err != nil {
		return err
	}
	defer store.Close()
	svc := service.New(store, fetcher.DeterministicClient())
	text, err := svc.InspectBundle(*id)
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}

func versionString() string { return "offlinebundle/1" }
