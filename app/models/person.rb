# frozen_string_literal: true

# == Schema Information
#
# Table name: people
#
#  id                   :bigint           not null, primary key
#  aasm_state           :string           default("published"), not null
#  birthday             :date
#  blood_type           :string
#  casts_count          :integer          default(0), not null
#  deleted_at           :datetime
#  favorite_users_count :integer          default(0), not null
#  gender               :string
#  height               :integer
#  name                 :string           not null
#  name_en              :string           default(""), not null
#  name_kana            :string           default(""), not null
#  nickname             :string
#  nickname_en          :string           default(""), not null
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
#  prefecture_id        :bigint
#
# Indexes
#
#  index_people_on_aasm_state            (aasm_state)
#  index_people_on_casts_count           (casts_count)
#  index_people_on_deleted_at            (deleted_at)
#  index_people_on_favorite_users_count  (favorite_users_count)
#  index_people_on_name                  (name) UNIQUE
#  index_people_on_prefecture_id         (prefecture_id)
#  index_people_on_staffs_count          (staffs_count)
#  index_people_on_unpublished_at        (unpublished_at)
#
# Foreign Keys
#
#  fk_rails_...  (prefecture_id => prefectures.id)
#

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

  after_save :touch_children
  after_destroy :touch_children

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

  private

  def touch_children
    casts.each(&:touch)
    staffs.each(&:touch)
  end
end
