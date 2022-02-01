package Controllers

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"time"

	Minio "Iagon/platform/minio"

	"github.com/minio/minio-go/v7"

	"Iagon/app/models"         // new
	config "Iagon/pkg/configs" // new

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson" // new
	"go.mongodb.org/mongo-driver/mongo"
	// new
	// new
)

var nonceSize int
var gcm cipher.AEAD
var encryptionSting = os.Getenv("MINIO_SECRETKEY")

const MySecret string = "abc&1*~#^2^#s0^=)^^7%b34"

var key = []byte(MySecret)

// var bytes_block, err = aes.NewCipher(key[:16])

// var bytess = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

// This should be in an env file in production

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(plainText []byte, MySecret string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return []byte{}, err
	}
	cfb := cipher.NewCFBEncrypter(block, key[:16])
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return cipherText, nil
}

// Decrypt method is to extract back the encrypted text
func Decrypt(cipherText []byte, MySecret string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return []byte{}, err
	}
	// cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, key[:16])
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return plainText, nil
}

func UploadFile(c *fiber.Ctx) error {
	ctx := context.Background()
	file, err := c.FormFile("fileUpload")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}
	if c.FormValue("objectName") == "" || c.FormValue("userId") == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "objectName or userId is missing",
		})
	}
	// Get Buffer from file
	buffer, err := file.Open()
	fileData, err := ioutil.ReadAll(buffer)
	if err != nil {
		return err
	}
	encryptedFile, err := Encrypt(fileData, MySecret)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}
	defer buffer.Close()

	// Create minio connection.
	minioClient, err := Minio.MinioConnection()
	if err != nil {
		// Return status 500 and minio connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}
	userObject, err := CreateUserObject(c)
	if err != nil {

		// Return status 500 and minio connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}
	objectName := file.Filename
	fileBuffer := bytes.NewReader(encryptedFile)
	contentType := file.Header["Content-Type"][0]
	fileSize := file.Size
	// Upload the zip file with PutObject
	Minio.CreateBucket(userObject.UserId)
	info, err := minioClient.PutObject(ctx, userObject.UserId, userObject.ObjectName, fileBuffer, fileSize, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"info":  info,
	})
}

func Download(c *fiber.Ctx) error {
	ctx := context.Background()

	if c.Query("objectName") == "" || c.Query("userId") == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "objectName or userId is missing",
		})
	}

	// Create minio connection.
	minioClient, err := Minio.MinioConnection()
	if err != nil {
		// Return status 500 and minio connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	err = UpdateUserObject(c)
	if err != nil {
		// Return status 500 and minio connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}
	// Upload the zip file with PutObject
	info, err := minioClient.GetObject(ctx, c.Query("userId"), c.Query("objectName"), minio.GetObjectOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}
	fileData, err := ioutil.ReadAll(info)
	// fmt.Println(info)
	// fmt.Println(fileData)

	plaintext, err := Decrypt(fileData, MySecret)
	// print(err.Error())
	// fmt.Println(plaintext)
	c.Response().Header.Set("Content-Type", "image/png")
	return c.SendStream(bytes.NewReader(plaintext))
}

func CreateUserObject(c *fiber.Ctx) (*models.Objects, error) {

	ObjectsCollection := config.MI.DB.Collection(os.Getenv("OBJECT_COLLECTION"))

	// find parameter
	data := new(models.Objects)

	// var data Request

	data.UserId = c.FormValue("userId")
	data.ObjectName = c.FormValue("objectName")
	data.DownloadedCount = 0
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	result, err := ObjectsCollection.InsertOne(c.Context(), data)

	if err != nil {
		return nil, err
	}

	// get the inserted data
	Objects := &models.Objects{}
	query := bson.D{{Key: "_id", Value: result.InsertedID}}

	ObjectsCollection.FindOne(c.Context(), query).Decode(Objects)

	return Objects, nil
}

func UpdateUserObject(c *fiber.Ctx) error {

	ObjectsCollection := config.MI.DB.Collection(os.Getenv("OBJECT_COLLECTION"))
	userId := c.Query("userId")
	objectName := c.Query("objectName")

	query := bson.M{"userid": userId, "objectname": objectName}

	update := bson.M{
		"$inc": bson.M{"downloadedcount": 1},
	}
	err := ObjectsCollection.FindOneAndUpdate(c.Context(), query, update).Err()
	// fmt.Println(result.UpsertedID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return err
		}

		return err
	}

	return nil
}
