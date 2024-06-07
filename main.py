import pika
import cv2
import numpy as np
from PIL import Image
import io

def process_image(image_data, filename):
    # Convert image data to a numpy array
    image_array = np.frombuffer(image_data, np.uint8)
    img = cv2.imdecode(image_array, cv2.IMREAD_COLOR)

    if img is None:
        print("Failed to decode the image")
        return ""

    # Calculate new dimensions while maintaining aspect ratio
    original_height, original_width = img.shape[:2]
    max_size = 200
    if original_width > original_height:
        new_width = max_size
        new_height = int((original_height * max_size) / original_width)
    else:
        new_height = max_size
        new_width = int((original_width * max_size) / original_height)

    # Resize the image
    resized_img = cv2.resize(img, (new_width, new_height), interpolation=cv2.INTER_AREA)

    # Determine output file path
    image = Image.open(io.BytesIO(image_data))
    file_extension = image.format.lower()
    success, encoded_img = cv2.imencode(f".{file_extension}", resized_img)
    if not success:
        print("Failed to encode image")
        return b""
    return encoded_img.tobytes()


def callback(ch, method, properties, body):
    filename = properties.headers["filename"]
    print(f"Received file: {filename}")

    out_name = process_image(body, filename)
    print(f"Compressed and resized {filename}")

    ch.basic_publish(
        exchange="",
        routing_key=properties.reply_to,
        properties=pika.BasicProperties(
            correlation_id=properties.correlation_id,
        ),
        body=out_name,
    )
    ch.basic_ack(delivery_tag=method.delivery_tag)


def main():
    connection_params = pika.ConnectionParameters("localhost")
    connection = pika.BlockingConnection(connection_params)
    channel = connection.channel()

    channel.queue_declare(queue="workers", durable=False)

    channel.basic_consume(queue="workers", on_message_callback=callback, auto_ack=False)

    print(" [*] Waiting for messages. To exit press CTRL+C")
    channel.start_consuming()


if __name__ == "__main__":
    main()
