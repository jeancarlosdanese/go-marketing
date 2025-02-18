// File: /internal/dto/import_contact.go

package dto

// ConfigImportContactDTO define as instruções para como a AI deve agir na importação de contatos
type ConfigImportContactDTO struct {
	Name       FieldMapping `json:"name"`       // Nome do contato
	Email      FieldMapping `json:"email"`      // E-mail do contato
	WhatsApp   FieldMapping `json:"whatsapp"`   // Número de telefone para WhatsApp
	Gender     FieldMapping `json:"gender"`     // Gênero do contato
	BirthDate  FieldMapping `json:"birth_date"` // Data de nascimento no formato "YYYY-MM-DD"
	Bairro     FieldMapping `json:"bairro"`     // Bairro onde reside
	Cidade     FieldMapping `json:"cidade"`     // Cidade onde reside
	Estado     FieldMapping `json:"estado"`     // Sigla do estado (UF)
	Interesses FieldMapping `json:"interesses"` // Como a AI deve categorizar os interesses
	Perfil     FieldMapping `json:"perfil"`     // Como a AI deve definir o perfil
	Eventos    FieldMapping `json:"eventos"`    // Como a AI deve categorizar os eventos
	History    FieldMapping `json:"history"`    // Como a AI deve gerar o histórico
}

// FieldMapping define como um campo do CSV deve ser interpretado pela IA
type FieldMapping struct {
	Source string            `json:"source"` // Nome da coluna no CSV que contém essa informação
	Rules  map[string]string `json:"rules"`  // Regras para a IA interpretar esse campo
}

// Exemplo de configuração JSON
const ExemploConfigJSON = `{
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
			"fallback": "Se fone_celular não existir, verificar se fone_residencial é um número de celular válido para WhatsApp"
		} 
	},
	"gender": { 
    "source": "sexo", 
		"rules": { 
			"mapping": "MASCULINO -> masculino, FEMININO -> feminino, OUTRO -> outro",
			"validate": "Se o nome indicar um gênero diferente, deixar em branco"
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
			"categorize": "Associar cursos feitos aos eventos"
		} 
	},
	"interesses": { 
		"source": "cursos", 
		"rules": { 
			"map": "Relacionar cursos a interesses [marketing_vendas, tecnologia_da_informacao, design_multimidia, tecnicas_profissionais_manutencao, outros_cursos, saude_bem_estar, idiomas, negocios_administracao_financas, desenvolvimento_pessoal_profissional, artesanato_moda, beleza_estetica, gastronomia_culinaria]"
		} 
	},
	"perfil": { 
		"source": "todos_os_campos", 
		"rules": { 
			"generate": "Utilizando as informações conhecidas descreva o perfil do contato" 
		} 
	},
	"history": { 
		"source": "todos_os_campos", 
		"rules": { 
			"generate": "Com base nos dados do registro, gerar um breve histórico do contato, incluindo cursos concluídos, experiências anteriores e áreas de interesse, mantendo um máximo de 250 caracteres"
		} 
	}
}`
