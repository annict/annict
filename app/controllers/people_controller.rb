# frozen_string_literal: true
# == Schema Information
#
# Table name: people
#
#  id               :integer          not null, primary key
#  prefecture_id    :integer
#  name             :string           not null
#  name_kana        :string           default(""), not null
#  nickname         :string
#  gender           :string
#  url              :string
#  wikipedia_url    :string
#  twitter_username :string
#  birthday         :date
#  blood_type       :string
#  height           :integer
#  aasm_state       :string           default("published"), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_people_on_aasm_state     (aasm_state)
#  index_people_on_name           (name) UNIQUE
#  index_people_on_prefecture_id  (prefecture_id)
#

class PeopleController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def show
    @person = Person.published.find(params[:id])

    if @person.voice_actor?
      @casts_with_year = @person.
        casts.
        published.
        joins(:work).
        where(works: { aasm_state: :published }).
        includes(:character, work: :work_image).
        group_by { |cast| cast.work.season_year.presence || 0 }
      @cast_years = @casts_with_year.keys.sort.reverse
    end

    if @person.staff?
      @staffs_with_year = @person.
        staffs.
        published.
        joins(:work).
        where(works: { aasm_state: :published }).
        includes(work: :work_image).
        group_by { |staff| staff.work.season_year.presence || 0 }
      @staff_years = @staffs_with_year.keys.sort.reverse
    end

    @favorite_people = @person.
      favorite_people.
      includes(user: :profile).
      joins(:user).
      merge(User.published).
      order(id: :desc)
  end

  private

  def load_i18n
    keys = {
      "messages._components.favorite_button.add_to_favorites": nil,
      "messages._components.favorite_button.added_to_favorites": nil
    }

    load_i18n_into_gon keys
  end
end
