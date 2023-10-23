package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kotsmile/codecrafters-bittorrent-go/torrent"
)

const (
	PeerId = "00112233445566778899"
	Port   = 6881
)

func DecodeSubcommand(bencodedValue string) (string, error) {
	decoded, _, err := torrent.DecodeBencode(bencodedValue)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(decoded)
	if err != nil {
		return "", err
	}

	return string(jsonOutput), nil
}

func InfoSubcommand(metaFilePath string) (string, error) {
	meta, err := torrent.NewMetaFromFile(metaFilePath)
	if err != nil {
		return "", err
	}

	metaInfo, err := meta.GetInfo()
	if err != nil {
		return "", err
	}

	var output string

	output += fmt.Sprintf("Tracker URL: %s\n", metaInfo.URL)
	output += fmt.Sprintf("Length: %d\n", metaInfo.Length)
	output += fmt.Sprintf("Info Hash: %s\n", metaInfo.InfoHash)
	output += fmt.Sprintf("Piece Length: %d\n", metaInfo.PieceLength)
	output += "Piece Hashes:"

	for _, piece := range metaInfo.PieceHashes {
		output += fmt.Sprintf("\n%s", piece)
	}

	return output, nil
}

func PeersSubcommand(metaFilePath string) (string, error) {
	meta, err := torrent.NewMetaFromFile(metaFilePath)
	if err != nil {
		return "", err
	}

	client := torrent.NewClient(meta, &torrent.Config{
		PeerId: PeerId,
		Port:   Port,
	})

	peersResponse, err := client.RequestPeers(0, 0, meta.Info.Length, 1)
	if err != nil {
		return "", err
	}

	var output string
	for _, peer := range peersResponse.Peers {
		output += fmt.Sprintf("\n%s", peer)
	}

	return output[1:], nil
}

func HandshakeSubcommand(metaFilePath, peerAddress string) (string, error) {
	meta, err := torrent.NewMetaFromFile(metaFilePath)
	if err != nil {
		return "", err
	}

	client := torrent.NewClient(meta, &torrent.Config{
		PeerId: PeerId,
		Port:   Port,
	})

	if err := client.Dial(peerAddress); err != nil {
		return "", err
	}
	defer client.Close(peerAddress)

	peerId, err := client.Handshake(peerAddress)
	if err != nil {
		return "", err
	}

	return peerId, nil
}

func DownloadPieceSubcommand(outputFilePath, metaFilePath string, pieceIndex int) ([]byte, error) {
	meta, err := torrent.NewMetaFromFile(metaFilePath)
	if err != nil {
		return nil, err
	}

	client := torrent.NewClient(meta, &torrent.Config{
		PeerId: PeerId,
		Port:   Port,
	})

	fmt.Println("Retrieve peers...")
	peersResponse, err := client.RequestPeers(0, 0, meta.Info.Length, 1)
	if err != nil {
		return nil, err
	}

	peerAddress := peersResponse.Peers[1]
	return client.ConnectAndGetDownloadPiece(peerAddress, pieceIndex)
}

func DownloadFileSubcommand(outputFilePath, metaFilePath string) ([]byte, error) {

	meta, err := torrent.NewMetaFromFile(metaFilePath)
	if err != nil {
		return nil, err
	}

	client := torrent.NewClient(meta, &torrent.Config{
		PeerId: PeerId,
		Port:   Port,
	})

	fmt.Println("Retrieve peers...")
	peersResponse, err := client.RequestPeers(0, 0, meta.Info.Length, 1)
	if err != nil {
		return nil, err
	}

	peerAddress := peersResponse.Peers[1]

	pieces, err := client.Meta.GetPieces()
	if err != nil {
		return nil, err
	}

	resultData := make([]byte, 0)

	for pieceIndex := range pieces {
		data, err := client.ConnectAndGetDownloadPiece(peerAddress, pieceIndex)
		if err != nil {
			return nil, err
		}

		resultData = append(resultData, data...)
	}

	return resultData, nil
}

func main() {
	subcommand := os.Args[1]

	switch subcommand {
	case "decode":
		output, err := DecodeSubcommand(os.Args[2])
		if err != nil {
			panic(err)
		}

		fmt.Println(output)
	case "info":
		torrentMetaFilePath := os.Args[2]
		output, err := InfoSubcommand(torrentMetaFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Println(output)
	case "peers":
		torrentMetaFilePath := os.Args[2]
		output, err := PeersSubcommand(torrentMetaFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Println(output)
	case "handshake":
		torrentMetaFilePath := os.Args[2]

		peerAddress := os.Args[3]

		output, err := HandshakeSubcommand(torrentMetaFilePath, peerAddress)
		if err != nil {
			panic(err)
		}

		fmt.Println("Peer ID:", output)
	case "download_piece":
		outputFilePath := os.Args[3]
		torrentMetaFilePath := os.Args[4]
		pieceIndex, err := strconv.Atoi(os.Args[5])
		if err != nil {
			panic(err)
		}

		output, err := DownloadPieceSubcommand(outputFilePath, torrentMetaFilePath, pieceIndex)
		if err != nil {
			panic(err)
		}

		file, err := os.Create(outputFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		file.Write(output)
		fmt.Printf("Piece %d downloaded to %s\n", pieceIndex, outputFilePath)
	case "download":
		outputFilePath := os.Args[3]
		torrentMetaFilePath := os.Args[4]

		data, err := DownloadFileSubcommand(outputFilePath, torrentMetaFilePath)
		if err != nil {
			panic(err)
		}

		file, err := os.Create(outputFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		file.Write(data)
		fmt.Printf("Downloaded %s to %s.\n", torrentMetaFilePath, outputFilePath)
	default:
		panic(fmt.Errorf("unknown subcommand: %s", subcommand))
	}
}
