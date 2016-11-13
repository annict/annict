# frozen_string_literal: true
# == Schema Information
#
# Table name: characters
#
#  id                 :integer          not null, primary key
#  name               :string           not null
#  name_en            :string           default(""), not null
#  kind               :string           default(""), not null
#  kind_en            :string           default(""), not null
#  nickname           :string           default(""), not null
#  nickname_en        :string           default(""), not null
#  birthday           :string           default(""), not null
#  birthday_en        :string           default(""), not null
#  age                :string           default(""), not null
#  age_en             :string           default(""), not null
#  blood_type         :string           default(""), not null
#  blood_type_en      :string           default(""), not null
#  height             :string           default(""), not null
#  height_en          :string           default(""), not null
#  weight             :string           default(""), not null
#  weight_en          :string           default(""), not null
#  nationality        :string           default(""), not null
#  nationality_en     :string           default(""), not null
#  occupation         :string           default(""), not null
#  occupation_en      :string           default(""), not null
#  description        :text             default(""), not null
#  description_en     :text             default(""), not null
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#  name_kana          :string           default(""), not null
#  aasm_state         :string           default("published"), not null
#  poster_image_id    :integer
#  cover_image_id     :integer
#  character_image_id :integer
#
# Indexes
#
#  index_characters_on_character_image_id  (character_image_id)
#  index_characters_on_cover_image_id      (cover_image_id)
#  index_characters_on_name_and_kind       (name,kind) UNIQUE
#  index_characters_on_poster_image_id     (poster_image_id)
#

class Character < ApplicationRecord
  include AASM
  include DbActivityMethods
  include RootResourceCommon

  DIFF_FIELDS = %i(
    name name_en kind kind_en nickname nickname_en birthday birthday_en age age_en
    blood_type blood_type_en height height_en weight weight_en nationality nationality_en
    occupation occupation_en description description_en name_kana
  ).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :character_image
  has_many :character_images, dependent: :destroy
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  validates :name, presence: true, uniqueness: { scope: :kind }

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
