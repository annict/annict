# frozen_string_literal: true

class OrganizationFansController < ApplicationV6Controller
  before_action :load_i18n, only: %i[index]

  def index
    set_page_category PageCategory::ORGANIZATION_FAN_LIST

    @organization = Organization.only_kept.find(params[:organization_id])
    @organization_favorites = @organization
      .organization_favorites
      .eager_load(user: :profile)
      .merge(User.only_kept)
      .order(watched_works_count: :desc)
  end

  private

  def load_i18n
    keys = {
      "verb.follow": nil,
      "noun.following": nil,
      "messages._components.favorite_button.add_to_favorites": nil,
      "messages._components.favorite_button.added_to_favorites": nil
    }

    load_i18n_into_gon keys
  end
end
