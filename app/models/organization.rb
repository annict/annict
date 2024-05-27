# frozen_string_literal: true

# == Schema Information
#
# Table name: organizations
#
#  id                   :bigint           not null, primary key
#  aasm_state           :string           default("published"), not null
#  deleted_at           :datetime
#  favorite_users_count :integer          default(0), not null
#  name                 :string           not null
#  name_en              :string           default(""), not null
#  name_kana            :string           default(""), not null
#  staffs_count         :integer          default(0), not null
#  twitter_username     :string
#  twitter_username_en  :string           default(""), not null
#  unpublished_at       :datetime
#  url                  :string
#  url_en               :string           default(""), not null
#  wikipedia_url        :string
#  wikipedia_url_en     :string           default(""), not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#
# Indexes
#
#  index_organizations_on_aasm_state            (aasm_state)
#  index_organizations_on_deleted_at            (deleted_at)
#  index_organizations_on_favorite_users_count  (favorite_users_count)
#  index_organizations_on_name                  (name) UNIQUE
#  index_organizations_on_staffs_count          (staffs_count)
#  index_organizations_on_unpublished_at        (unpublished_at)
#

class Organization < ApplicationRecord
  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[
    name name_kana url wikipedia_url twitter_username name_kana name_en url_en
    wikipedia_url_en twitter_username_en
  ].freeze

  validates :name, presence: true, uniqueness: true
  validates :url, url: {allow_blank: true}
  validates :url_en, url: {allow_blank: true}
  validates :wikipedia_url, url: {allow_blank: true}
  validates :wikipedia_url_en, url: {allow_blank: true}

  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  # organization_favorites are user data. so do not add `dependent: :destroy`
  has_many :organization_favorites
  has_many :staffs, as: :resource, dependent: :destroy
  has_many :staff_works, source: :work, through: :staffs
  has_many :users, through: :organization_favorites

  def favorites
    organization_favorites
  end

  after_destroy :touch_children
  after_save :touch_children

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  private

  def touch_children
    staffs.each(&:touch)
  end
end
