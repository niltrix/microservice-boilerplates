package main

import (
	"context"
	"io/ioutil"

	// "context"
	"fmt"
	"log"
	"os"

	// fiber2
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	// open telemetry
	/*	"go.opentelemetry.io/otel"
		"go.opentelemetry.io/otel/attribute"
		stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
		"go.opentelemetry.io/otel/propagation"
		"go.opentelemetry.io/otel/sdk/resource"
		sdktrace "go.opentelemetry.io/otel/sdk/trace"
		semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
		oteltrace "go.opentelemetry.io/otel/trace"
	*/
	"github.com/signalfx/splunk-otel-go/distro"

	// oauth2
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2/clientcredentials"
)

type Subscription struct {
	Name    string `json:"name"`
	Product string `json:"product"`
}

func getSubscription(c *fiber.Ctx) error {
	checkppid()
	subscription := Subscription{
		Name:    "Elon",
		Product: "Tesla",
	}
	return c.Status(fiber.StatusOK).JSON(subscription)
}

func createSubscription(c *fiber.Ctx) error {
	checkppid()
	subs := new(Subscription)
	err := c.BodyParser(subs)
	if err != nil {
		err := c.Status(fiber.StatusBadRequest).SendString(err.Error())
		if err != nil {
			fmt.Printf("Error : [%s]", err)
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(subs)
}

func main() {
	os.Setenv("HTTPS_PROXY", "web-proxy.sg.hpicorp.net:8080")
	//os.Setenv("HTTPS_PROXY", "15.89.14.62:8080")

	_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=my-app,service.version=1.2.3,deployment.environment=development")
	sdk, err := distro.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	provider, err := oidc.NewProvider(context.Background(), "https://pie.authz.wpp.api.hp.com/")

	if err != nil {
		fmt.Printf("oidc error : %s", err)
	}
	fmt.Println(provider.Endpoint().AuthURL)

	// Configure an OpenID Connect aware OAuth2 client.
	/*oauth2Config := oauth2.Config{
		ClientID:     "fdsafdsa",
		ClientSecret: "fdsafdsaf",
		RedirectURL:  nil,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}*/
	clientConfig := clientcredentials.Config{
		ClientID:     "f730cae7fe0e430793bde8b2d46fcf95",
		ClientSecret: "09d1fbf0a3d64525bcd1b3cb0103de2b",
		TokenURL:     provider.Endpoint().TokenURL,
	}
	httpClient := clientConfig.Client(context.Background())
	response, err := httpClient.Get("https://stratus-pie.tropos-rnd.com/v2/tenantmgtsvc/tenants/c1a46987-cbd3-45db-b94a-a6544116fc2d")
	//response, err := httpClient.Post("http://stratus-pie.tropos-rnd.com/v2/tenantmgtsvc/tenants/", "application/json", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	tenants, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer response.Body.Close()

	log.Println(string(tenants))
	token, token_error := clientConfig.Token(context.Background())
	if token_error != nil {
		log.Fatal(token_error.Error())
	}
	log.Println(token.AccessToken)

	app := fiber.New(fiber.Config{
		//Prefork: true,
	})
	app.Use(logger.New())
	app.Use(requestid.New(requestid.Config{
		Header: "x-request-id",
	}))

	app.Get("/", func(ctx *fiber.Ctx) error {
		// _, span := tracer.Start(ctx.UserContext(), "getUser", oteltrace.WithAttributes(attribute.String("id", ctx.BaseURL())))
		// defer span.End()
		checkppid()
		return ctx.SendString("Hello")
	})

	app.Get("/subscription", getSubscription)
	app.Post("/subscription", createSubscription)

	err_app := app.Listen(":8080")
	if err_app != nil {
		fmt.Printf("Error : [%s]", err)
	}
}

// func initTracer() *sdktrace.TracerProvider {
// 	exporter, err := stdout.New(stdout.WithPrettyPrint())
// 	//exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	tp := sdktrace.NewTracerProvider(
// 		sdktrace.WithSampler(sdktrace.AlwaysSample()),
// 		sdktrace.WithBatcher(exporter),
// 		sdktrace.WithResource(
// 			resource.NewWithAttributes(
// 				semconv.SchemaURL,
// 				semconv.ServiceNameKey.String("my-service"),
// 			)),
// 	)
// 	otel.SetTracerProvider(tp)
// 	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
// 	return tp
// }

// Print current process
func checkppid() {
	if fiber.IsChild() {
		fmt.Printf("[%d] Child\n", os.Getppid())
	} else {
		fmt.Printf("[%d] Master\n", os.Getppid())
	}
}
