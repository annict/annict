# typed: false
# frozen_string_literal: true

class Character < ApplicationRecord
  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[
    name name_en series_id nickname nickname_en birthday birthday_en age age_en
    blood_type blood_type_en height height_en weight weight_en nationality nationality_en
    occupation occupation_en description description_en name_kana description_source
    description_source_en
  ].freeze

  belongs_to :series
  has_many :casts, dependent: :destroy
  has_many :works, through: :casts
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :character_favorites
  has_many :users, through: :character_favorites

  validates :series_id, presence: true
  validates :name, presence: true, uniqueness: {scope: :series_id}
  validates :description, presence_pair: :description_source
  validates :description_en, presence_pair: :description_source_en

  def favorites
    character_favorites
  end

  def oldest_work
    works.order("season_year ASC, season_name ASC").first
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end
end
