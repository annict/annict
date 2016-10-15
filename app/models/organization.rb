# frozen_string_literal: true
# == Schema Information
#
# Table name: organizations
#
#  id                  :integer          not null, primary key
#  name                :string           not null
#  url                 :string
#  wikipedia_url       :string
#  twitter_username    :string
#  aasm_state          :string           default("published"), not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  name_kana           :string           default(""), not null
#  name_en             :string           default(""), not null
#  url_en              :string           default(""), not null
#  wikipedia_url_en    :string           default(""), not null
#  twitter_username_en :string           default(""), not null
#
# Indexes
#
#  index_organizations_on_aasm_state  (aasm_state)
#  index_organizations_on_name        (name) UNIQUE
#

class Organization < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include OrganizationCommon

  validates :name, uniqueness: true

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      after do
        staffs.published.each(&:hide!)
      end

      transitions from: :published, to: :hidden
    end
  end

  has_many :draft_staffs, as: :resource, dependent: :destroy
  has_many :staffs, as: :resource, dependent: :destroy
end
