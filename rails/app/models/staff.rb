# typed: false
# frozen_string_literal: true

class Staff < ApplicationRecord
  extend Enumerize

  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i[
    resource_id name role role_other sort_number name_en role_other_en
  ].freeze

  enumerize :role, in: %w[
    original_creator
    chief_director
    director
    series_composition
    script
    original_character_design
    character_design
    chief_animation_director
    animation_director
    art_director
    sound_director
    music
    studio
    other
  ]

  counter_culture :resource, column_name: ->(staff) { staff.published? ? :staffs_count : nil }

  belongs_to :resource, polymorphic: true
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  validates :resource, presence: true
  validates :work_id, presence: true
  validates :name, presence: true
  validates :role, presence: true

  scope :major, -> { where.not(role: "other") }

  before_validation :set_name

  localized_method :accurate_name, :name

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = case field
      when :role
        send(field).to_s if send(field).present?
      else
        send(field)
      end

      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def support_en?
    name_en.present? && resource.name_en.present?
  end

  def accurate_name
    return name if name == resource.name
    "#{name} (#{resource.name})"
  end

  def accurate_name_en
    return name_en if name_en == resource.name_en
    "#{name_en} (#{resource.name_en})"
  end

  def local_name_with_old
    return local_name if local_name == resource.local_name
    "#{local_name} (#{resource.local_name})"
  end

  def role_name
    return local_role_other if role_value == "other"
    role_text
  end

  def local_role_other
    return role_other_en if I18n.locale != :ja && role_other_en.present?
    role_other
  end

  private

  def set_name
    self.name = resource.name if name.blank? && resource.present?
    self.name_en = resource.name_en if name_en.blank? && resource&.name_en.present?
  end
end
