package incident

import "github.com/slack-go/slack"

func createIncidentModal() slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "Create an Incident", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Cancel", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Create Incident", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Enter the below details for the incident channel.", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	statusText := slack.NewTextBlockObject("plain_text", "status", false, false)
	statusPlaceholder := slack.NewTextBlockObject("plain_text", "Current Status...", false, false)
	statusOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "Investigating", false, false), slack.NewTextBlockObject("plain_text", "We think something is wrong, but we're not sure what it is yet", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "Fixing", false, false), slack.NewTextBlockObject("plain_text", "We know what is wrong and we are trying to fix it.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "Monitoring", false, false), slack.NewTextBlockObject("plain_text", "We think it is fixed, but we are waiting for confirmation.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "Resolved", false, false), slack.NewTextBlockObject("plain_text", "We fixed it.", false, false)),
	}
	statusSelection := slack.NewOptionsSelectBlockElement("static_select", statusPlaceholder, "status", statusOptions...)
	status := slack.NewInputBlock("status", statusText, nil, statusSelection)

	severityText := slack.NewTextBlockObject("plain_text", "severity", false, false)
	severityPlaceholder := slack.NewTextBlockObject("plain_text", "Select Severity...", false, false)
	severityOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "SEV-1", false, false), slack.NewTextBlockObject("plain_text", "Patient care is blocked.", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "SEV-2", false, false), slack.NewTextBlockObject("plain_text", "Patient care is critically degraded.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "SEV-3", false, false), slack.NewTextBlockObject("plain_text", "Risk to patient care, but no current impact.", false, false)),
		slack.NewOptionBlockObject("4", slack.NewTextBlockObject("plain_text", "SEV-4", false, false), slack.NewTextBlockObject("plain_text", "Internal incident, no risk to patient care.", false, false)),
	}
	severitySelection := slack.NewOptionsSelectBlockElement("static_select", severityPlaceholder, "incident_severity", severityOptions...)
	severity := slack.NewInputBlock("incident_severity", severityText, nil, severitySelection)

	descriptionText := slack.NewTextBlockObject("plain_text", "description", false, false)
	descriptionHint := slack.NewTextBlockObject("plain_text", "Example: Increased latency in Carehub", false, false)
	decriptionPlaceholder := slack.NewTextBlockObject("plain_text", "Enter description of incident...", false, false)
	descriptionElement := slack.NewPlainTextInputBlockElement(decriptionPlaceholder, "description")
	description := slack.NewInputBlock("description", descriptionText, descriptionHint, descriptionElement)

	incidentMembersText := slack.NewTextBlockObject("plain_text", "incident_members", false, false)
	incidentMembersHint := slack.NewTextBlockObject("plain_text", "Example: @johndoe @janedoe", false, false)
	incidentMembersPlaceholder := slack.NewTextBlockObject("plain_text", "Enter people to add to incident channel...", false, false)
	incidentMembersSelection := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, incidentMembersPlaceholder, "incident_members")
	incidentMembers := slack.NewInputBlock("incident_members", incidentMembersText, incidentMembersHint, incidentMembersSelection)
	incidentMembers.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			severity,
			status,
			description,
			incidentMembers,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	modalRequest.CallbackID = "create_incident_modal"

	return modalRequest
}

func updateIncidentModal() slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "Update an Incident", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Cancel", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Update Incident", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Enter the below details for the incident update.", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	statusText := slack.NewTextBlockObject("plain_text", "status", false, false)
	statusPlaceholder := slack.NewTextBlockObject("plain_text", "Current Status...", false, false)
	statusOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "Investigating", false, false), slack.NewTextBlockObject("plain_text", "We think something is wrong, but we're not sure what it is yet", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "Fixing", false, false), slack.NewTextBlockObject("plain_text", "We know what is wrong and we are trying to fix it.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "Monitoring", false, false), slack.NewTextBlockObject("plain_text", "We think it is fixed, but we are waiting for confirmation.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "Resolved", false, false), slack.NewTextBlockObject("plain_text", "We fixed it.", false, false)),
	}
	statusSelection := slack.NewOptionsSelectBlockElement("static_select", statusPlaceholder, "status", statusOptions...)
	status := slack.NewInputBlock("status", statusText, nil, statusSelection)

	severityText := slack.NewTextBlockObject("plain_text", "severity", false, false)
	severityPlaceholder := slack.NewTextBlockObject("plain_text", "Select Severity...", false, false)
	severityOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "SEV-1", false, false), slack.NewTextBlockObject("plain_text", "Patient care is blocked.", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "SEV-2", false, false), slack.NewTextBlockObject("plain_text", "Patient care is critically degraded.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "SEV-3", false, false), slack.NewTextBlockObject("plain_text", "Risk to patient care, but no current impact.", false, false)),
		slack.NewOptionBlockObject("4", slack.NewTextBlockObject("plain_text", "SEV-4", false, false), slack.NewTextBlockObject("plain_text", "Internal incident, no risk to patient care.", false, false)),
	}
	severitySelection := slack.NewOptionsSelectBlockElement("static_select", severityPlaceholder, "incident_severity", severityOptions...)
	severity := slack.NewInputBlock("incident_severity", severityText, nil, severitySelection)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			severity,
			status,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	modalRequest.CallbackID = "update_incident_modal"

	return modalRequest
}
