# frozen_string_literal: true

# == Schema Information
#
# Table name: casts
#
#  id             :bigint           not null, primary key
#  aasm_state     :string           default("published"), not null
#  deleted_at     :datetime
#  name           :string           not null
#  name_en        :string           default(""), not null
#  sort_number    :integer          default(0), not null
#  unpublished_at :datetime
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#  anime_id       :bigint           not null
#  character_id   :bigint           not null
#  person_id      :bigint           not null
#
# Indexes
#
#  index_casts_on_aasm_state      (aasm_state)
#  index_casts_on_anime_id        (anime_id)
#  index_casts_on_character_id    (character_id)
#  index_casts_on_deleted_at      (deleted_at)
#  index_casts_on_person_id       (person_id)
#  index_casts_on_sort_number     (sort_number)
#  index_casts_on_unpublished_at  (unpublished_at)
#
# Foreign Keys
#
#  fk_rails_...  (anime_id => animes.id)
#  fk_rails_...  (character_id => characters.id)
#  fk_rails_...  (person_id => people.id)
#

class Cast < ApplicationRecord
  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i(person_id name part sort_number character_id name_en).freeze

  counter_culture :person, column_name: -> (cast) { cast.published? ? :casts_count : nil }

  belongs_to :character, touch: true
  belongs_to :person, touch: true
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  validates :character_id, presence: true
  validates :name, presence: true
  validates :person_id, presence: true
  validates :work_id, presence: true

  before_validation :set_name

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field) if respond_to?(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  def support_en?
    name_en.present? && character.name_en.present? && person.name_en.present?
  end

  private

  def set_name
    self.name = person.name if name.blank? && person.present?
    self.name_en = person.name_en if name_en.blank? && person&.name_en.present?
  end
end
