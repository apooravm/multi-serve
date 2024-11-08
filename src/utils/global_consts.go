package utils

import "os"

var (
	PORT string
	// files
	REQUEST_LOG_PATH string
	SERVER_LOG_PATH  string

	CLIPBOARD_PATH_JSON string
	CLIPBOARD_PATH_TXT  string

	DUMMY_WS_LOG_PATH string

	CHAT_DEBUG string
	CHAT_LOG   string
	// ---
	CHAT_PASS string

	BUCKET_NAME         string
	BUCKET_REGION       string
	OBJ_RESUME_KEY      string
	OBJ_RESUME_PNG_KEY  string
	OBJ_RESUME_HTML_KEY string
	OBJ_RESUME_MD_KEY   string

	NOTES_DATA_FOLDER      string
	LOCAL_INFO_PATH        string
	LOCAL_RESUME_PATH      string
	LOCAL_RESUME_PNG_PATH  string
	LOCAL_RESUME_HTML_PATH string
	LOCAL_RESUME_MD_PATH   string

	QUERY_GENERAL_PASS string
	QUERY_TRIGGER_PASS string

	DB_URL     string
	DB_API_KEY string
)

func InitGlobalVars() {
	PORT = os.Getenv("PORT")
	// REQUEST_LOG_PATH = os.Getenv("REQUEST_LOG_PATH")
	REQUEST_LOG_PATH = "./data/logs/request_logs.json"

	// SERVER_LOG_PATH = os.Getenv("SERVER_LOG_PATH")
	SERVER_LOG_PATH = "./data/logs/server_logs.log"

	DUMMY_WS_LOG_PATH = "./data/logs/dummy_ws.log"

	// CHAT_DEBUG = os.Getenv("CHAT_DEBUG")
	CHAT_DEBUG = "./data/logs/chat_debug.log"
	// CHAT_LOG = os.Getenv("CHAT_LOG")
	CHAT_LOG = "./data/logs/chat_log.log"

	CHAT_PASS = os.Getenv("CHAT_PASS")

	BUCKET_NAME = os.Getenv("BUCKET_NAME")
	BUCKET_REGION = os.Getenv("BUCKET_REGION")
	NOTES_DATA_FOLDER = os.Getenv("NOTES_DATA_FOLDER")

	LOCAL_INFO_PATH = "./data/notes/notesinfo.json"

	// Resume in different formats.
	LOCAL_RESUME_PATH = "./data/S3/Apoorav_Medal_CV.pdf"
	LOCAL_RESUME_PNG_PATH = "./data/S3/Apoorav_Medal_CV.png"
	LOCAL_RESUME_HTML_PATH = "./data/S3/Apoorav_Medal_CV.html"
	LOCAL_RESUME_MD_PATH = "./data/S3/Apoorav_Medal_CV.md"

	// S3 object keys to fetch resume.
	OBJ_RESUME_KEY = os.Getenv("OBJ_RESUME_KEY")
	OBJ_RESUME_PNG_KEY = os.Getenv("OBJ_RESUME_PNG_KEY")
	OBJ_RESUME_HTML_KEY = os.Getenv("OBJ_RESUME_HTML_KEY")
	OBJ_RESUME_MD_KEY = os.Getenv("OBJ_RESUME_MD_KEY")

	// Param auth passwords
	QUERY_GENERAL_PASS = os.Getenv("QUERY_GENERAL_PASS")
	QUERY_TRIGGER_PASS = os.Getenv("QUERY_TRIGGER_PASS")

	DB_URL = os.Getenv("DB_URL")
	DB_API_KEY = os.Getenv("DB_KEY")

	CLIPBOARD_PATH_JSON = "./data/WebClipboard.json"
	CLIPBOARD_PATH_TXT = "./data/WebClipboard.txt"
}
