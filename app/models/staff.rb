# frozen_string_literal: true
# == Schema Information
#
# Table name: staffs
#
#  id            :integer          not null, primary key
#  work_id       :integer          not null
#  name          :string           not null
#  role          :string           not null
#  role_other    :string
#  aasm_state    :string           default("published"), not null
#  sort_number   :integer          default(0), not null
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#  resource_id   :integer          not null
#  resource_type :string           not null
#  name_en       :string           default(""), not null
#  role_other_en :string           default(""), not null
#
# Indexes
#
#  index_staffs_on_aasm_state                     (aasm_state)
#  index_staffs_on_resource_id_and_resource_type  (resource_id,resource_type)
#  index_staffs_on_sort_number                    (sort_number)
#  index_staffs_on_work_id                        (work_id)
#

class Staff < ApplicationRecord
  extend Enumerize
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(
    resource_id name role role_other sort_number name_en role_other_en
  ).freeze

  enumerize :role, in: %w(
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
  )

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :resource, polymorphic: true, counter_cache: true
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  validates :resource, presence: true
  validates :work_id, presence: true
  validates :name, presence: true
  validates :role, presence: true

  scope :major, -> { where.not(role: "other") }

  before_validation :set_name

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :role
        send(field).to_s if send(field).present?
      else
        send(field)
      end

      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  private

  def set_name
    self.name = resource.name if name.blank? && resource.present?
    self.name_en = resource.name_en if name_en.blank? && resource&.name_en.present?
  end
end
