# frozen_string_literal: true
# == Schema Information
#
# Table name: characters
#
#  id             :integer          not null, primary key
#  name           :string           not null
#  name_en        :string           default(""), not null
#  kind           :string           default(""), not null
#  kind_en        :string           default(""), not null
#  nickname       :string           default(""), not null
#  nickname_en    :string           default(""), not null
#  birthday       :string           default(""), not null
#  birthday_en    :string           default(""), not null
#  age            :string           default(""), not null
#  age_en         :string           default(""), not null
#  blood_type     :string           default(""), not null
#  blood_type_en  :string           default(""), not null
#  height         :string           default(""), not null
#  height_en      :string           default(""), not null
#  weight         :string           default(""), not null
#  weight_en      :string           default(""), not null
#  nationality    :string           default(""), not null
#  nationality_en :string           default(""), not null
#  occupation     :string           default(""), not null
#  occupation_en  :string           default(""), not null
#  description    :text             default(""), not null
#  description_en :text             default(""), not null
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#  name_kana      :string           default(""), not null
#  aasm_state     :string           default("published"), not null
#
# Indexes
#
#  index_characters_on_name_and_kind  (name,kind) UNIQUE
#

class Character < ApplicationRecord
  include AASM
  include DbActivityMethods
  include RootResourceCommon

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  validates :name, presence: true, uniqueness: { scope: :kind }
end
