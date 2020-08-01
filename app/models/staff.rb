# frozen_string_literal: true
# == Schema Information
#
# Table name: staffs
#
#  id             :bigint           not null, primary key
#  aasm_state     :string           default("published"), not null
#  deleted_at     :datetime
#  name           :string           not null
#  name_en        :string           default(""), not null
#  resource_type  :string           not null
#  role           :string           not null
#  role_other     :string
#  role_other_en  :string           default(""), not null
#  sort_number    :integer          default(0), not null
#  unpublished_at :datetime
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#  anime_id       :bigint           not null
#  resource_id    :bigint           not null
#
# Indexes
#
#  index_staffs_on_aasm_state                     (aasm_state)
#  index_staffs_on_anime_id                       (anime_id)
#  index_staffs_on_deleted_at                     (deleted_at)
#  index_staffs_on_resource_id_and_resource_type  (resource_id,resource_type)
#  index_staffs_on_sort_number                    (sort_number)
#  index_staffs_on_unpublished_at                 (unpublished_at)
#
# Foreign Keys
#
#  fk_rails_...  (anime_id => animes.id)
#

class Staff < ApplicationRecord
  extend Enumerize

  include DbActivityMethods
  include Unpublishable

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

  counter_culture :resource, column_name: -> (staff) { staff.published? ? :staffs_count : nil }

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

  def support_en?
    name_en.present? && resource.name_en.present?
  end

  private

  def set_name
    self.name = resource.name if name.blank? && resource.present?
    self.name_en = resource.name_en if name_en.blank? && resource&.name_en.present?
  end
end
