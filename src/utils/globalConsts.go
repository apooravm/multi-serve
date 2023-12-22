package utils

import "os"

var (
	PORT             string
	REQUEST_LOG_PATH string

	CHAT_DEBUG string
	CHAT_LOG   string
	CHAT_PASS  string

	BUCKET_NAME    string
	BUCKET_REGION  string
	OBJ_RESUME_KEY string

	NOTES_DATA_FOLDER string
	LOCAL_INFO_PATH   string
	LOCAL_RESUME_PATH string

	QUERY_GENERAL_PASS string
	QUERY_TRIGGER_PASS string

	DB_URL     string
	DB_API_KEY string
)

func InitGlobalVars() {
	PORT = os.Getenv("PORT")
	REQUEST_LOG_PATH = os.Getenv("REQUEST_LOG_PATH")

	CHAT_DEBUG = os.Getenv("CHAT_DEBUG")
	CHAT_LOG = os.Getenv("CHAT_LOG")
	CHAT_PASS = os.Getenv("CHAT_PASS")

	BUCKET_NAME = os.Getenv("BUCKET_NAME")
	BUCKET_REGION = os.Getenv("BUCKET_REGION")
	NOTES_DATA_FOLDER = os.Getenv("NOTES_DATA_FOLDER")

	LOCAL_INFO_PATH = os.Getenv("LOCAL_INFO_PATH")

	QUERY_TRIGGER_PASS = os.Getenv("QUERY_TRIGGER_PASS")

	LOCAL_RESUME_PATH = os.Getenv("LOCAL_RESUME_PATH")

	OBJ_RESUME_KEY = os.Getenv("OBJ_RESUME_KEY")

	QUERY_GENERAL_PASS = os.Getenv("QUERY_GENERAL_PASS")

	DB_URL = os.Getenv("DB_URL")
	DB_API_KEY = os.Getenv("DB_KEY")
}
