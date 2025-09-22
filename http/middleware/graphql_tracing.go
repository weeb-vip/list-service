package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/weeb-vip/list-service/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GraphQLTracingExtension provides OpenTelemetry tracing for GraphQL operations
type GraphQLTracingExtension struct{}

// ExtensionName returns the name of the extension
func (e GraphQLTracingExtension) ExtensionName() string {
	return "GraphQLTracing"
}

// Validate validates the extension configuration
func (e GraphQLTracingExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptOperation traces GraphQL operations
func (e GraphQLTracingExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	rc := graphql.GetOperationContext(ctx)

	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "GraphQL "+string(rc.Operation.Operation),
		trace.WithAttributes(
			attribute.String("graphql.operation.name", rc.OperationName),
			attribute.String("graphql.operation.type", string(rc.Operation.Operation)),
			attribute.String("graphql.document", rc.RawQuery),
		),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// Log GraphQL operation start
	log.Ctx(ctx).Info().
		Str("operation_name", rc.OperationName).
		Str("operation_type", string(rc.Operation.Operation)).
		Msg("GraphQL operation started")

	responseHandler := next(ctx)

	return func(ctx context.Context) *graphql.Response {
		response := responseHandler(ctx)

		// Log any GraphQL errors
		if response.Errors != nil && len(response.Errors) > 0 {
			for _, err := range response.Errors {
				span.RecordError(err)
				log.Ctx(ctx).Error().
					Err(err).
					Str("operation_name", rc.OperationName).
					Msg("GraphQL operation error")
			}
		}

		log.Ctx(ctx).Info().
			Str("operation_name", rc.OperationName).
			Str("operation_type", string(rc.Operation.Operation)).
			Msg("GraphQL operation completed")

		return response
	}
}

// InterceptField traces individual GraphQL field resolutions
func (e GraphQLTracingExtension) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)

	tracer := tracing.GetTracer(ctx)
	ctx, span := tracer.Start(ctx, "GraphQL Field: "+fc.Field.Name,
		trace.WithAttributes(
			attribute.String("graphql.field.name", fc.Field.Name),
			attribute.String("graphql.field.object", fc.Field.ObjectDefinition.Name),
			attribute.String("graphql.field.type", fc.Field.Definition.Type.String()),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	result, err := next(ctx)

	if err != nil {
		span.RecordError(err)
		log.Ctx(ctx).Error().
			Err(err).
			Str("field_name", fc.Field.Name).
			Str("object_type", fc.Field.ObjectDefinition.Name).
			Msg("GraphQL field resolution error")
	}

	return result, err
}