// File: /internal/service/sqs_service.go

package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

// SQSService define as operações para interagir com o Amazon SQS
type SQSService interface {
	SendMessage(ctx context.Context, queueName string, message interface{}) error
	ReceiveMessages(ctx context.Context, queueName string, handler func(msg dto.CampaignMessageDTO) error) error
}

// sqsService gerencia a comunicação com o Amazon SQS
type sqsService struct {
	log              *slog.Logger
	client           *sqs.Client
	emailQueueURL    string
	whatsappQueueURL string
}

// NewSQSService inicializa o serviço de filas SQS
func NewSQSService(emailQueueURL, whatsappQueueURL string) (*sqsService, error) {
	log := logger.GetLogger()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Error("Erro ao carregar configuração AWS", "error", err)
		return nil, err
	}

	client := sqs.NewFromConfig(cfg)

	return &sqsService{
		log:              log,
		client:           client,
		emailQueueURL:    emailQueueURL,
		whatsappQueueURL: whatsappQueueURL,
	}, nil
}

// SendMessage envia uma mensagem para a fila correta (email ou whatsapp)
func (s *sqsService) SendMessage(ctx context.Context, queueName string, message interface{}) error {
	var queueURL string

	if queueName == "email" {
		queueURL = s.emailQueueURL
	} else if queueName == "whatsapp" {
		queueURL = s.whatsappQueueURL
	} else {
		s.log.Warn("Tipo de fila inválido", "queueName", queueName)
		return nil
	}

	// 🚀 Serializa apenas uma vez
	messageBody, err := json.Marshal(message)
	if err != nil {
		s.log.Error("Erro ao serializar mensagem para SQS", "error", err)
		return err
	}

	// 🔥 Envia sem converter para string (evita JSON aninhado)
	_, err = s.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(messageBody)), // <---- Mantém JSON puro
	})
	if err != nil {
		s.log.Error("Erro ao enviar mensagem para SQS", "queue", queueName, "error", err)
		return err
	}

	s.log.Info("Mensagem enviada para SQS com sucesso", "queue", queueName)
	return nil
}

// ReceiveMessages processa mensagens de uma fila específica e passa para um handler
func (s *sqsService) ReceiveMessages(ctx context.Context, queueName string, handler func(msg dto.CampaignMessageDTO) error) error {
	var queueURL string

	if queueName == "email" {
		queueURL = s.emailQueueURL
	} else if queueName == "whatsapp" {
		queueURL = s.whatsappQueueURL
	} else {
		s.log.Warn("Tipo de fila inválido", "queueName", queueName)
		return nil
	}

	for {
		msgResult, err := s.client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     10,
		})
		if err != nil {
			s.log.Error("Erro ao receber mensagens do SQS", "queue", queueName, "error", err)
			continue
		}

		for _, message := range msgResult.Messages {
			s.log.Info("📩 Mensagem recebida do SQS", "queue", queueName, "message_id", *message.MessageId)

			// 🔍 Remover aspas extras caso existam
			var rawMessage *string
			if err := json.Unmarshal([]byte(*message.Body), &rawMessage); err == nil {
				s.log.Debug("Mensagem estava aninhada, extraindo JSON real...")
			} else {
				rawMessage = message.Body
			}

			// 🔄 Desserializar JSON da mensagem diretamente para a estrutura correta
			var campaignMessage dto.CampaignMessageDTO
			if err := json.Unmarshal([]byte(*rawMessage), &campaignMessage); err != nil {
				s.log.Error("Erro ao decodificar mensagem do SQS", "error", err, "raw_msg", rawMessage)
				continue
			}

			s.log.Info("📦 Mensagem decodificada com sucesso", "queue", queueName, "contact_id", campaignMessage.ContactID)
			// 🚀 Chama a função de processamento do worker
			if err := handler(campaignMessage); err != nil {
				s.log.Error("Erro ao processar mensagem", "error", err)
				continue
			}

			// 🗑️ Remover mensagem da fila após processamento
			_, err := s.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				s.log.Error("Erro ao deletar mensagem do SQS", "queue", queueName, "error", err)
			}
		}
	}
}
