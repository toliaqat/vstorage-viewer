# VStorage Viewer

This is a terminal-based application for traversing a tree-like structure of data from the Agoric VStorage API. The application displays the data in a series of columns and provides navigation using arrow keys. Additionally, it includes a data panel for displaying detailed JSON data.

## Features

- Traverse a tree-like structure of data from the Agoric VStorage API.
- Display children nodes in a series of columns.
- Navigate between columns using left and right arrow keys.
- Get decoded data from the selected node.

## Install dependencies
This project uses Go modules. Ensure you have Go installed and then run:

```shell
go mod tidy
```

## Usage

Run the application:

```shell
go run main.go
```
