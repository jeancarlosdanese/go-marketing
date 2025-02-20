// File: /internal/dto/import_contact.go

package dto

// ConfigImportContactDTO define as instruções para como a AI deve agir na importação de contatos
type ConfigImportContactDTO struct {
	AboutData     FieldMapping `json:"about_data"`      // Informações gerais sobre os dados
	Name          FieldMapping `json:"name"`            // Nome do contato
	Email         FieldMapping `json:"email"`           // E-mail do contato
	WhatsApp      FieldMapping `json:"whatsapp"`        // Número de telefone para WhatsApp
	Gender        FieldMapping `json:"gender"`          // Gênero do contato
	BirthDate     FieldMapping `json:"birth_date"`      // Data de nascimento no formato "YYYY-MM-DD"
	Bairro        FieldMapping `json:"bairro"`          // Bairro onde reside
	Cidade        FieldMapping `json:"cidade"`          // Cidade onde reside
	Estado        FieldMapping `json:"estado"`          // Sigla do estado (UF)
	Interesses    FieldMapping `json:"interesses"`      // Como a AI deve categorizar os interesses
	Perfil        FieldMapping `json:"perfil"`          // Como a AI deve definir o perfil
	Eventos       FieldMapping `json:"eventos"`         // Como a AI deve categorizar os eventos
	History       FieldMapping `json:"history"`         // Como a AI deve gerar o histórico
	LastContactAt FieldMapping `json:"last_contact_at"` // Como a AI deve definir a última data de contato
}

// FieldMapping define como um campo do CSV deve ser interpretado pela IA
type FieldMapping struct {
	Source string            `json:"source"` // Nome da coluna no CSV que contém essa informação
	Rules  map[string]string `json:"rules"`  // Regras para a IA interpretar esse campo
}

// Exemplo de configuração JSON
const ExemploConfigJSON = `
{
	"about_data": {
		"source": "todos_os_campos",
		"rules": {
			"info": "Dados de alunos que concluíram cursos de capacitação profissional livre em nossa escola..."
		}
	},
	"name": {
		"source": "nome",
		"rules": {
			"default": "Utilizar o nome do contato"
		}
	},
	"email": {
		"source": "email",
		"rules": {
			"default": "Utilizar o e-mail do contato"
		}
	},
	"whatsapp": {
		"source": "fone_celular",
		"rules": {
			"fallback": "Se fone_celular não existir, verificar fone_residencial"
		}
	},
	"gender": {
		"source": "nome",
		"rules": {
			"check_name": "Definir gênero pelo nome, se não for possível deixar vazio."
		}
	},
	"birth_date": {
		"source": "data_nascimento",
		"rules": {
			"format": "YYYY-MM-DD"
		}
	},
	"bairro": {
		"source": "bairro",
		"rules": {}
	},
	"cidade": {
		"source": "cidade",
		"rules": {}
	},
	"estado": {
		"source": "uf",
		"rules": {}
	},
	"eventos": {
		"source": "cursos",
		"rules": {
			"categorize": "Associar cursos concluídos aos eventos relevantes"
		}
	},
	"interesses": {
		"source": "cursos",
		"rules": {
			"map": "Relacionar cursos a interesses em áreas específicas (possible_tags)",
			"possible_tags": "marketing_vendas, tecnologia_da_informacao, design_multimidia, tecnicas_profissionais_manutencao, saude_bem_estar, idiomas, negocios_gestao_financas, desenvolvimento_pessoal_profissional, artesanato_moda, beleza_estetica, gastronomia_culinaria"
		}
	},
	"perfil": {
		"source": "profissao,local_trabalho",
		"rules": {
			"categorize": "Categorize o perfil (possible_tags) com base nas informações de profissão e local de trabalho",
			"possible_tags": "industria, producao, construcao_civil, manutencao, logistica, comercial, tecnologia, saude_bem_estar, educacao, financas, gestao, marketing, seguranca, engenharia, juridico, agronegocio, meio_ambiente"
		}
	},
	"history": {
		"source": "todos_os_campos",
		"rules": {
			"generate": "Criar um breve histórico do aluno com base nas informações disponíveis. Deve retornar um texto de até 500 caracteres, garantindo que: 
1. O nome do aluno esteja sempre corretamente capitalizado, no formato 'João da Silva' (não todo maiúsculo, nem todo minúsculo). 
2. Datas devem ser formatadas como 'DD/MM/YYYY'. 
3. O texto deve ser relatar apenas fatos de forma natural e profissional, evitando repetições desnecessárias."
		}
	},
	"last_contact_at": {
		"source": "cursos",
		"rules": {
			"default_source": "Utilizar última data de conclusão de cursos. Format 'YYYY-MM-DD'"
		}
	}
}
	`
