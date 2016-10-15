# frozen_string_literal: true
# == Schema Information
#
# Table name: casts
#
#  id           :integer          not null, primary key
#  person_id    :integer          not null
#  work_id      :integer          not null
#  name         :string           not null
#  part         :string           not null
#  aasm_state   :string           default("published"), not null
#  sort_number  :integer          default(0), not null
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#  character_id :integer
#  name_en      :string           default(""), not null
#
# Indexes
#
#  index_casts_on_aasm_state                (aasm_state)
#  index_casts_on_character_id              (character_id)
#  index_casts_on_person_id                 (person_id)
#  index_casts_on_sort_number               (sort_number)
#  index_casts_on_work_id                   (work_id)
#  index_casts_on_work_id_and_character_id  (work_id,character_id) UNIQUE
#

class Cast < ActiveRecord::Base
  include AASM
  include DbActivityMethods

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :character
  belongs_to :person
  belongs_to :work, touch: true

  validates :character_id, presence: true
  validates :name, presence: true
  validates :person_id, presence: true
  validates :work_id, presence: true

  before_validation :set_name

  private

  def set_name
    self.name = person.name if name.blank? && person.present?
    self.name_en = person.name_en if name_en.blank? && person&.name_en.present?
  end
end
