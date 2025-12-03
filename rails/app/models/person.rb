# typed: false
# frozen_string_literal: true

class Person < ApplicationRecord
  extend Enumerize

  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[
    prefecture_id name name_kana nickname gender url wikipedia_url twitter_username
    birthday blood_type height name_en nickname_en url_en wikipedia_url_en
    twitter_username_en
  ].freeze

  enumerize :blood_type, in: %i[a b ab o]
  enumerize :gender, in: %i[male female]

  validates :name, presence: true, uniqueness: true
  validates :url, url: {allow_blank: true}
  validates :url_en, url: {allow_blank: true}
  validates :wikipedia_url, url: {allow_blank: true}
  validates :wikipedia_url_en, url: {allow_blank: true}

  belongs_to :prefecture, optional: true
  has_many :casts, dependent: :destroy
  has_many :cast_works, source: :work, through: :casts
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  # person_favorites are user data. so do not add `dependent: :destroy`
  has_many :person_favorites
  has_many :staffs, as: :resource, dependent: :destroy
  has_many :staff_works, source: :work, through: :staffs
  has_many :users, through: :person_favorites

  def favorites
    person_favorites
  end

  after_destroy :touch_children
  after_save :touch_children

  def voice_actor?
    casts.exists?
  end

  def staff?
    staffs.exists?
  end

  def attributes=(params)
    super
    self.birthday = Date.parse(params[:birthday]) if params[:birthday].present?
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def self.ransackable_attributes(auth_object = nil)
    %w[name name_en name_kana]
  end

  def self.ransackable_associations(auth_object = nil)
    []
  end

  private

  def touch_children
    casts.each(&:touch)
    staffs.each(&:touch)
  end
end
