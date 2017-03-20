# frozen_string_literal: true

class PersonFansController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index(person_id)
    @person = Person.published.find(person_id)
    @favorite_people = @person.
      favorite_people.
      order(watched_works_count: :desc)
  end

  private

  def load_i18n
    keys = {
      "verb.follow": nil,
      "noun.following": nil,
      "messages.components.favorite_button.add_to_favorites": nil,
      "messages.components.favorite_button.added_to_favorites": nil
    }

    load_i18n_into_gon keys
  end
end
