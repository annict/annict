# frozen_string_literal: true
# == Schema Information
#
# Table name: people
#
#  id                   :integer          not null, primary key
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
#  url                  :string
#  url_en               :string           default(""), not null
#  wikipedia_url        :string
#  wikipedia_url_en     :string           default(""), not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#  prefecture_id        :integer
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
#
# Foreign Keys
#
#  fk_rails_...  (prefecture_id => prefectures.id)
#

class Person < ApplicationRecord
  extend Enumerize

  include DBActivityMethods
  include RootResourceCommon
  include SoftDeletable

  DIFF_FIELDS = %i(
    prefecture_id name name_kana nickname gender url wikipedia_url twitter_username
    birthday blood_type height name_en nickname_en url_en wikipedia_url_en
    twitter_username_en
  ).freeze

  enumerize :blood_type, in: %i(a b ab o)
  enumerize :gender, in: %i(male female)

  validates :name, presence: true, uniqueness: true
  validates :url, url: { allow_blank: true }
  validates :url_en, url: { allow_blank: true }
  validates :wikipedia_url, url: { allow_blank: true }
  validates :wikipedia_url_en, url: { allow_blank: true }

  belongs_to :prefecture, optional: true
  has_many :casts, dependent: :destroy
  has_many :cast_works, through: :casts, source: :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :favorite_people, dependent: :destroy
  has_many :staffs, as: :resource, dependent: :destroy
  has_many :staff_works, through: :staffs, source: :work
  has_many :users, through: :favorite_people

  def favorites
    favorite_people
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
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  def soft_delete_with_children
    soft_delete
    casts.without_deleted.each(&:soft_delete)
    staffs.without_deleted.each(&:soft_delete)
  end

  private

  def touch_children
    casts.each(&:touch)
    staffs.each(&:touch)
  end
end
