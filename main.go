package main

import (
	"bytes"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"workers", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			filename := d.Headers["filename"].(string)
			id := d.Headers["id"]
			_ = id
			// fmt.Println("Received file:", filename, "ID:", id)
			fmt.Println("Received file")

			outName := processImage(d.Body, filename)

			log.Printf("Compressed and resized %s", filename)

			// Send the response to the callback queue
			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key (callback queue)
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					Body:          []byte(outName),
					CorrelationId: d.CorrelationId,
				})
			failOnError(err, "Failed to publish a message")

			// Acknowledge the original message
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func processImage(imageData []byte, filename string) string {

	// Create a bytes reader.
	reader := bytes.NewReader(imageData)

	// Decode the image.
	img, format, err := image.Decode(reader)
	if err != nil {
		log.Printf("Failed to encode and save image: %v", err)
        return ""
	}

	// Get the original dimensions.
	originalBounds := img.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio.
	var newWidth, newHeight int
	maxSize := 200
	if originalWidth > originalHeight {
		newWidth = maxSize
		newHeight = (originalHeight * maxSize) / originalWidth
	} else {
		newHeight = maxSize
		newWidth = (originalWidth * maxSize) / originalHeight
	}

	// Resize the image.
	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, originalBounds, draw.Over, nil)

	// Determine the output file name and compression settings based on the format.
	var outFileName string
	var outFile *os.File
	var encoderError error

	switch format {
	case "jpeg":
		// s := strconv.Itoa(rand.Int())
		outFileName = fmt.Sprintf("../chatapp-api/images/%s.jpg", filename)
		outFile, err = os.Create(outFileName)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer outFile.Close()
		// Set the JPEG compression quality.
		jpegOptions := jpeg.Options{Quality: 80} // Adjust quality as needed.
		encoderError = jpeg.Encode(outFile, resizedImg, &jpegOptions)

	case "png":
		// s := strconv.Itoa(rand.Int())
		outFileName = fmt.Sprintf("../chatapp-api/images/%s.png", filename)
		outFile, err = os.Create(outFileName)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer outFile.Close()
		// Set the PNG compression level.
		pngEncoder := png.Encoder{CompressionLevel: png.BestCompression} // Adjust level as needed.
		encoderError = pngEncoder.Encode(outFile, resizedImg)

	default:
		log.Fatalf("Unsupported image format: %v", format)
	}

	if encoderError != nil {
		log.Printf("Failed to encode and save image: %v", encoderError)
        outFileName = ""
	}

	return outFileName

	// outBytes, err := fileToByteArray(outFile)
	// if err != nil {
	// 	panic(err)
	// }

}

func fileToByteArray(file *os.File) ([]byte, error) {
	// Ensure the file offset is at the beginning
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Read the file into a byte slice
	byteArray, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return byteArray, nil
}
