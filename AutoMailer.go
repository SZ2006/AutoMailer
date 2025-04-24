package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"os"
	"path/filepath"

	gomail "gopkg.in/mail.v2" // Geef een alias om naamconflict te voorkomen
)

// Foutmeldingen als constante waarden
const (
	errTeWeinigParams     = "Te weinig parameters. Gebruik: automailer <bestandspad> <e-mailadres>"
	errBestandBestaatNiet = "Fout: opgegeven bestand bestaat niet."
	errOngeldigEmail      = "Fout: ongeldig e-mailadres."
	errLadenConfig        = "Fout bij inlezen configuratie."
	errVerzendenEmail     = "Fout bij verzenden van e-mail."
)

// Structuur voor de instellingen uit config.json
type Config struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

func main() {
	// Open of maak een logbestand aan
	logBestand, _ := os.OpenFile("automailer.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logBestand.Close()
	log.SetOutput(logBestand)

	if len(os.Args) < 3 {
		log.Println(errTeWeinigParams)
		fmt.Println(errTeWeinigParams)
		return
	}

	bestandPad := os.Args[1]
	emailadres := os.Args[2]

	if _, err := os.Stat(bestandPad); os.IsNotExist(err) {
		log.Printf("Bestand bestaat niet: %s\n", bestandPad)
		fmt.Println(errBestandBestaatNiet)
		return
	}

	if !isValidEmail(emailadres) {
		log.Printf("Ongeldig e-mailadres: %s\n", emailadres)
		fmt.Println(errOngeldigEmail)
		return
	}

	absConfigPad, _ := filepath.Abs("config.json")
	config, err := loadConfig(absConfigPad)
	if err != nil {
		log.Println("Fout bij laden van configuratie:", err)
		fmt.Println(errLadenConfig)
		return
	}

	err = sendMail(config, emailadres, bestandPad)
	if err != nil {
		log.Println("Fout bij verzenden:", err)
		fmt.Println(errVerzendenEmail)
		return
	}

	log.Printf("Succesvol verzonden naar %s: %s\n", emailadres, bestandPad)
	fmt.Println("E-mail succesvol verzonden.")
}

func loadConfig(configPath string) (*Config, error) {
	bestand, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("kan configuratiebestand niet openen: %w", err)
	}
	defer bestand.Close()

	var config Config
	decoder := json.NewDecoder(bestand)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("configuratiebestand is ongeldig: %w", err)
	}
	return &config, nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func sendMail(cfg *Config, recipient, filePath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", "Automatisch bestand verzonden")
	m.SetBody("text/plain", "Zie bijlage")
	m.Attach(filePath)

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.Username, cfg.Password)
	return d.DialAndSend(m)
}
