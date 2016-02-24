# == Schema Information
#
# Table name: people
#
#  id               :integer          not null, primary key
#  prefecture_id    :integer
#  name             :string           not null
#  name_kana        :string
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
  def show(id)
    @person = Person.published.find(id)

    if @person.voice_actor?
      @casts_with_year = @person.
        casts.
        includes(work: [:season, :item]).
        group_by { |cast| cast.work.season&.year.presence || 0 }
      @cast_years = @casts_with_year.keys.sort.reverse
    end

    if @person.staff?
      @staffs_with_year = @person.
        staffs.
        includes(work: [:season, :item]).
        group_by { |staff| staff.work.season&.year.presence || 0 }
      @staff_years = @staffs_with_year.keys.sort.reverse
    end
  end
end
