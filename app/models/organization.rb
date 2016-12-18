# frozen_string_literal: true
# == Schema Information
#
# Table name: organizations
#
#  id               :integer          not null, primary key
#  name             :string           not null
#  url              :string
#  wikipedia_url    :string
#  twitter_username :string
#  aasm_state       :string           default("published"), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#  name_kana        :string           default(""), not null
#
# Indexes
#
#  index_organizations_on_aasm_state  (aasm_state)
#  index_organizations_on_name        (name) UNIQUE
#

class Organization < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include RootResourceCommon

  DIFF_FIELDS = %i(
    name name_kana url wikipedia_url twitter_username name_kana name_en url_en
    wikipedia_url_en twitter_username_en
  ).freeze

  validates :name, presence: true, uniqueness: true
  validates :url, url: { allow_blank: true }
  validates :url_en, url: { allow_blank: true }
  validates :wikipedia_url, url: { allow_blank: true }
  validates :wikipedia_url_en, url: { allow_blank: true }

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

  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :staffs, as: :resource, dependent: :destroy

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
