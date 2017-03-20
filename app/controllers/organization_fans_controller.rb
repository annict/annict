# frozen_string_literal: true

class OrganizationFansController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index(organization_id)
    @organization = Organization.published.find(organization_id)
    @favorite_orgs = @organization.
      favorite_organizations.
      order(watched_works_count: :desc)
  end

  private

  def load_i18n
    keys = {
      "verb.follow": nil,
      "noun.following": nil,
      "messages.components.favorite_button.add_to_favorites": nil,
      "messages.components.favorite_button.added_to_favorites": nil,
    }

    load_i18n_into_gon keys
  end
end
