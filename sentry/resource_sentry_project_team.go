package sentry

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jianyuan/go-sentry/sentry"
	"strings"
)

func resourceSentryProjectTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSentryProjectTeamCreate,
		ReadContext:   resourceSentryProjectTeamRead,
		DeleteContext: resourceSentryProjectTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSentryProjectTeamImporter,
		},
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the organization the project belongs to",
				ForceNew:    true,
			},
			"team": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the team to create the project for",
				ForceNew:    true,
			},
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the organization the project belongs to",
				ForceNew:    true,
			},
		},
	}
}

func resourceSentryProjectTeamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*sentry.Client)

	org := d.Get("organization").(string)
	team := d.Get("team").(string)
	project := d.Get("project").(string)

	proj, _, err := client.Projects.AddTeam(org, project, team)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Add team to Sentry project", map[string]interface{}{
		"projectSlug": proj.Slug,
		"projectID":   proj.ID,
		"team":        team,
		"org":         org,
	})

	d.SetId(fmt.Sprintf("%s/%s", proj.Slug, team))
	d.Set("organization", org)
	d.Set("team", team)
	d.Set("project", project)
	return resourceSentryProjectRead(ctx, d, meta)
}

func resourceSentryProjectTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*sentry.Client)
	org := d.Get("organization").(string)

	var diags diag.Diagnostics

	identifier := strings.Split("/", d.Id())
	if len(identifier) != 2 {
		return diag.Errorf("unexpected identifier for a project team. expected format {project}/{team}")
	}

	project := identifier[0]
	team := identifier[1]

	proj, resp, err := client.Projects.Get(org, project)
	if found, err := checkClientGet(resp, err, d); !found {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Read Sentry project to see if team belongs to project", map[string]interface{}{
		"projectSlug": proj.Slug,
		"projectID":   proj.ID,
		"org":         org,
		"teams":       proj.Teams,
	})

	for _, t := range proj.Teams {
		if t.Slug == team {
			return diags
		}
	}

	// Remove the team from the state because it is not assigned to the Sentry project
	d.SetId("")

	return diags
}

func resourceSentryProjectTeamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*sentry.Client)

	org := d.Get("organization").(string)
	parts := strings.Split("/", d.Id())
	if len(parts) != 2 {
		return diag.Errorf("unexpected identifier for a project team. expected format {project}/{team}")
	}

	project := parts[0]
	team := parts[1]
	_, err := client.Projects.RemoveTeam(org, project, team)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Remove team from Sentry project", map[string]interface{}{
		"projectSlug": project,
		"teamSlug":    team,
		"org":         org,
	})

	return diag.FromErr(err)
}

func resourceSentryProjectTeamImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	addrID := d.Id()

	parts := strings.Split(addrID, "/")

	if len(parts) != 3 {
		return nil, errors.New("unexpected identifier for a project team. expected format org-slug/project-slug/team-slug")
	}

	tflog.Debug(ctx, "Importing Sentry project team", map[string]interface{}{
		"org":         parts[0],
		"projectSlug": parts[1],
		"teamSlug":    parts[2],
	})

	d.Set("organization", parts[0])
	d.SetId(fmt.Sprintf("%s/%s", parts[1], parts[2]))

	return []*schema.ResourceData{d}, nil
}
