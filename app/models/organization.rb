# typed: false
# frozen_string_literal: true

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
