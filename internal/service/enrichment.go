package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	repositories "custodian/internal/repository"
	"log"
)

var EnrichmentQueue chan models.Event

func StartEnrichmentWorker() {
	EnrichmentQueue = make(chan models.Event, 1000)

	go func() {
		for event := range EnrichmentQueue {
			profileRepo := repositories.NewProfileRepository(pkg.GetMongoDBInstance().Database, "profiles")

			// Step 1: Enrich
			if err := EnrichProfile(event); err != nil {
				log.Println("❌ Enrichment failed for:", event.PermaId, err)
				continue
			}

			// Step 2: Unify
			profile, err := profileRepo.FindProfileByID(event.PermaId)
			if err == nil && profile != nil {
				log.Println("🔄 Unifying profile:", profile.PermaId)
				if _, err := unifyProfiles(*profile); err != nil {
					log.Println("❌ Unification failed for:", profile.PermaId, err)
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
