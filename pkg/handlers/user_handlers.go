package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"lexichat-backend/pkg/models"
	jwt "lexichat-backend/pkg/utils/auth"
	"lexichat-backend/pkg/utils/logging"
	log "lexichat-backend/pkg/utils/logging"
	utils "lexichat-backend/pkg/utils/users"
)


func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.ErrorLogger.Printf("Error decoding JSON: %v\n", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		err = user.Create(db)
		if err != nil {
			log.ErrorLogger.Printf("Error creating user: %v\n", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		token, err := jwt.GenerateJWT(user.UserID)
		if err != nil {
			log.ErrorLogger.Printf("Error generating JWT token: %v\n", err)
			http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
			return
		}

		log.InfoLogger.Printf("User created successfully: %s\n", user.UserID)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"userId": string(user.UserID),
			"token":  token,
		})
	}
}


// user discovery
func DiscoverUsersByUserIdHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    partialUserID := r.URL.Query().Get("partialUserId")

    // Minimum partialUserID must exceed 3 characters
    if len(partialUserID) < 3 {
        http.Error(w, "partialUserId must be longer", http.StatusBadRequest)
        return
    }

    users, err := utils.DiscoverUsersByUserId(db, partialUserID)
    if err != nil {
		errorMessage := fmt.Sprintf("Problem arose while discovering users: %v", err)
        http.Error(w, errorMessage, http.StatusInternalServerError)
		logging.ErrorLogger.Fatalln(errorMessage)
		return
    }

    // Marshal users slice to JSON
    usersJSON, err := json.Marshal(users)
    if err != nil {
        http.Error(w, "Problem converting users to JSON", http.StatusInternalServerError)
        fmt.Println("Error marshaling users to JSON:", err)
		logging.ErrorLogger.Fatalln("Error marshaling users to JSON while discovering users. ", err)
        return
    }

    // Set Content-Type header and write JSON response
    w.Header().Set("Content-Type", "application/json")
    w.Write(usersJSON)
}

func FetchUserDetailsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, userID string) {
	var user models.User
	user, err := FetchUserDetails(userID, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Problem fetching user details: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func FetchUserDetails(userID string, db *sql.DB) (models.User, error){
	query := `SELECT user_id, user_name, phone_number, profile_picture, created_at, fcm_token FROM users WHERE user_id = $1`

	var user models.User
	err := db.QueryRow(query, userID).Scan(&user.UserID, &user.Username, &user.PhoneNumber, &user.ProfilePicture, &user.CreatedAt, &user.FCMToken)

	print(fmt.Printf("%#v", user))

	if err != nil {
		logging.ErrorLogger.Printf( fmt.Sprintf("Error in fetching user details. %s", err.Error()))
		return models.User{}, err
	}
	return user, nil
}