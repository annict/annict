# frozen_string_literal: true

# == Schema Information
#
# Table name: characters
#
#  id                    :bigint           not null, primary key
#  aasm_state            :string           default("published"), not null
#  age                   :string           default(""), not null
#  age_en                :string           default(""), not null
#  birthday              :string           default(""), not null
#  birthday_en           :string           default(""), not null
#  blood_type            :string           default(""), not null
#  blood_type_en         :string           default(""), not null
#  deleted_at            :datetime
#  description           :text             default(""), not null
#  description_en        :text             default(""), not null
#  description_source    :string           default(""), not null
#  description_source_en :string           default(""), not null
#  favorite_users_count  :integer          default(0), not null
#  height                :string           default(""), not null
#  height_en             :string           default(""), not null
#  name                  :string           not null
#  name_en               :string           default(""), not null
#  name_kana             :string           default(""), not null
#  nationality           :string           default(""), not null
#  nationality_en        :string           default(""), not null
#  nickname              :string           default(""), not null
#  nickname_en           :string           default(""), not null
#  occupation            :string           default(""), not null
#  occupation_en         :string           default(""), not null
#  unpublished_at        :datetime
#  weight                :string           default(""), not null
#  weight_en             :string           default(""), not null
#  created_at            :datetime         not null
#  updated_at            :datetime         not null
#  series_id             :bigint
#
# Indexes
#
#  index_characters_on_deleted_at            (deleted_at)
#  index_characters_on_favorite_users_count  (favorite_users_count)
#  index_characters_on_name_and_series_id    (name,series_id) UNIQUE
#  index_characters_on_series_id             (series_id)
#  index_characters_on_unpublished_at        (unpublished_at)
#
# Foreign Keys
#
#  fk_rails_...  (series_id => series.id)
#

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
