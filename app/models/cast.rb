# typed: false
# frozen_string_literal: true

class Cast < ApplicationRecord
  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i[person_id name part sort_number character_id name_en].freeze

  counter_culture :person, column_name: ->(cast) { cast.published? ? :casts_count : nil }

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

  localized_method :name

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field) if respond_to?(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def support_en?
    name_en.present? && character.name_en.present? && person.name_en.present?
  end

  def accurate_name
    return name if name == person.name
    "#{name} (#{person.name})"
  end

  def accurate_name_en
    return name_en if name_en == person.name_en
    "#{name_en} (#{person.name_en})"
  end

  def local_name_with_old
    return local_name if local_name == person.local_name
    "#{local_name} (#{person.local_name})"
  end

  private

  def set_name
    self.name = person.name if name.blank? && person.present?
    self.name_en = person.name_en if name_en.blank? && person&.name_en.present?
  end
end
