package logic

import (
	"context"
	"errors"
	"strings"

	"maestros/internal/svc"
	"maestros/maestros"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetPersonaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPersonaLogic(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
) *GetPersonaLogic {
	return &GetPersonaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPersonaLogic) GetPersona(
	in *maestros.GetPersonaRequest,
) (*maestros.GetPersonaResponse, error) {

	// Validar que el ID haya sido enviado.
	if in == nil || strings.TrimSpace(in.Id) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"el id de la persona es obligatorio",
		)
	}

	personaID := strings.TrimSpace(in.Id)

	// Consultar la persona en maestros_db.
	p, err := l.svcCtx.PersonaRepository.FindByID(
		l.ctx,
		personaID,
	)

	if err != nil {
		// La persona no existe.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(
				codes.NotFound,
				"persona no encontrada",
			)
		}

		// Error inesperado de PostgreSQL.
		l.Errorf(
			"error consultando persona %s: %v",
			personaID,
			err,
		)

		return nil, status.Error(
			codes.Internal,
			"no se pudo consultar la persona",
		)
	}

	// Convertir el modelo de PostgreSQL al mensaje gRPC.
	persona := &maestros.Persona{
		Id:              p.ID,
		TipoDocumento:   p.TipoDocumento,
		NumeroDocumento: p.NumeroDocumento,
		Nombres:         p.Nombres,
		ApellidoPaterno: p.ApellidoPaterno,
		ApellidoMaterno: p.ApellidoMaterno,
		Activo:          p.Activo,
		CreatedAt: p.CreatedAt.Format(
			"2006-01-02T15:04:05Z07:00",
		),
		UpdatedAt: p.UpdatedAt.Format(
			"2006-01-02T15:04:05Z07:00",
		),
	}

	// Campos opcionales.
	if p.FechaNacimiento != nil {
		persona.FechaNacimiento = p.FechaNacimiento.Format(
			"2006-01-02",
		)
	}

	if p.Sexo != nil {
		persona.Sexo = *p.Sexo
	}

	if p.CorreoPersonal != nil {
		persona.CorreoPersonal = *p.CorreoPersonal
	}

	if p.Telefono != nil {
		persona.Telefono = *p.Telefono
	}

	return &maestros.GetPersonaResponse{
		Persona: persona,
	}, nil
}
