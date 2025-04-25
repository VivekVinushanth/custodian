package service

import (
	"github.com/wso2/identity-customer-data-service/pkg/locks"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	repositories "github.com/wso2/identity-customer-data-service/pkg/repository"
	"log"
)

var EnrichmentQueue chan models.Event

func StartEnrichmentWorker() {
	EnrichmentQueue = make(chan models.Event, 1000)

	go func() {
		for event := range EnrichmentQueue {
			profileRepo := repositories.NewProfileRepository(locks.GetMongoDBInstance().Database, "profiles")

			// Step 1: Enrich
			if err := EnrichProfile(event); err != nil {
				log.Println("❌ Enrichment failed for:", event.ProfileId, err)
				continue
			}

			// Step 2: Unify
			profile, err := profileRepo.FindProfileByID(event.ProfileId)
			if err == nil && profile != nil {
				log.Println("🔄 Unifying profile:", profile.ProfileId)
				if _, err := unifyProfiles(*profile); err != nil {
					log.Println("❌ Unification failed for:", profile.ProfileId, err)
				}
			}
		}
	}()
}

func EnqueueEventForProcessing(event models.Event) {
	if EnrichmentQueue != nil {
		EnrichmentQueue <- event
	}
}
