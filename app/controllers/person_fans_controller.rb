# frozen_string_literal: true

class PersonFansController < ApplicationV6Controller
  before_action :load_i18n, only: %i[index]

  def index
    set_page_category PageCategory::PERSON_FAN_LIST

    @person = Person.only_kept.find(params[:person_id])
    @person_favorites = @person
      .person_favorites
      .joins(:user)
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
