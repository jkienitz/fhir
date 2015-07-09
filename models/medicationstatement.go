// Copyright (c) 2011-2015, HL7, Inc & The MITRE Corporation
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
//     * Redistributions of source code must retain the above copyright notice, this
//       list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above copyright notice,
//       this list of conditions and the following disclaimer in the documentation
//       and/or other materials provided with the distribution.
//     * Neither the name of HL7 nor the names of its contributors may be used to
//       endorse or promote products derived from this software without specific
//       prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package models

import "encoding/json"

type MedicationStatement struct {
	Id                          string                               `json:"-" bson:"_id"`
	Identifier                  []Identifier                         `bson:"identifier,omitempty" json:"identifier,omitempty"`
	Patient                     *Reference                           `bson:"patient,omitempty" json:"patient,omitempty"`
	InformationSource           *Reference                           `bson:"informationSource,omitempty" json:"informationSource,omitempty"`
	DateAsserted                *FHIRDateTime                        `bson:"dateAsserted,omitempty" json:"dateAsserted,omitempty"`
	Status                      string                               `bson:"status,omitempty" json:"status,omitempty"`
	WasNotTaken                 *bool                                `bson:"wasNotTaken,omitempty" json:"wasNotTaken,omitempty"`
	ReasonNotTaken              []CodeableConcept                    `bson:"reasonNotTaken,omitempty" json:"reasonNotTaken,omitempty"`
	ReasonForUseCodeableConcept *CodeableConcept                     `bson:"reasonForUseCodeableConcept,omitempty" json:"reasonForUseCodeableConcept,omitempty"`
	ReasonForUseReference       *Reference                           `bson:"reasonForUseReference,omitempty" json:"reasonForUseReference,omitempty"`
	EffectiveDateTime           *FHIRDateTime                        `bson:"effectiveDateTime,omitempty" json:"effectiveDateTime,omitempty"`
	EffectivePeriod             *Period                              `bson:"effectivePeriod,omitempty" json:"effectivePeriod,omitempty"`
	Note                        string                               `bson:"note,omitempty" json:"note,omitempty"`
	MedicationCodeableConcept   *CodeableConcept                     `bson:"medicationCodeableConcept,omitempty" json:"medicationCodeableConcept,omitempty"`
	MedicationReference         *Reference                           `bson:"medicationReference,omitempty" json:"medicationReference,omitempty"`
	Dosage                      []MedicationStatementDosageComponent `bson:"dosage,omitempty" json:"dosage,omitempty"`
}

type MedicationStatementDosageComponent struct {
	Text                    string           `bson:"text,omitempty" json:"text,omitempty"`
	Schedule                *Timing          `bson:"schedule,omitempty" json:"schedule,omitempty"`
	AsNeededBoolean         *bool            `bson:"asNeededBoolean,omitempty" json:"asNeededBoolean,omitempty"`
	AsNeededCodeableConcept *CodeableConcept `bson:"asNeededCodeableConcept,omitempty" json:"asNeededCodeableConcept,omitempty"`
	Site                    *CodeableConcept `bson:"site,omitempty" json:"site,omitempty"`
	Route                   *CodeableConcept `bson:"route,omitempty" json:"route,omitempty"`
	Method                  *CodeableConcept `bson:"method,omitempty" json:"method,omitempty"`
	Quantity                *Quantity        `bson:"quantity,omitempty" json:"quantity,omitempty"`
	Rate                    *Ratio           `bson:"rate,omitempty" json:"rate,omitempty"`
	MaxDosePerPeriod        *Ratio           `bson:"maxDosePerPeriod,omitempty" json:"maxDosePerPeriod,omitempty"`
}

type MedicationStatementBundle struct {
	Id    string                           `json:"id,omitempty"`
	Type  string                           `json:"resourceType,omitempty"`
	Base  string                           `json:"base,omitempty"`
	Total int                              `json:"total,omitempty"`
	Link  []BundleLinkComponent            `json:"link,omitempty"`
	Entry []MedicationStatementBundleEntry `json:"entry,omitempty"`
}

type MedicationStatementBundleEntry struct {
	Id       string                `json:"id,omitempty"`
	Base     string                `json:"base,omitempty"`
	Link     []BundleLinkComponent `json:"link,omitempty"`
	Resource MedicationStatement   `json:"resource,omitempty"`
}

func (resource *MedicationStatement) MarshalJSON() ([]byte, error) {
	x := struct {
		ResourceType string `json:"resourceType"`
		MedicationStatement
	}{
		ResourceType:        "MedicationStatement",
		MedicationStatement: *resource,
	}
	return json.Marshal(x)
}
