# frozen_string_literal: true
# == Schema Information
#
# Table name: people
#
#  id                    :integer          not null, primary key
#  prefecture_id         :integer
#  name                  :string           not null
#  name_kana             :string           default(""), not null
#  nickname              :string
#  gender                :string
#  url                   :string
#  wikipedia_url         :string
#  twitter_username      :string
#  birthday              :date
#  blood_type            :string
#  height                :integer
#  aasm_state            :string           default("published"), not null
#  created_at            :datetime         not null
#  updated_at            :datetime         not null
#  name_en               :string           default(""), not null
#  nickname_en           :string           default(""), not null
#  url_en                :string           default(""), not null
#  wikipedia_url_en      :string           default(""), not null
#  twitter_username_en   :string           default(""), not null
#  favorites_count       :integer          default(0), not null
#  favorite_people_count :integer          default(0), not null
#
# Indexes
#
#  index_people_on_aasm_state     (aasm_state)
#  index_people_on_name           (name) UNIQUE
#  index_people_on_prefecture_id  (prefecture_id)
#

class Person < ActiveRecord::Base
  extend Enumerize
  include AASM
  include DbActivityMethods
  include RootResourceCommon

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

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      after do
        casts.published.each(&:hide!)
        staffs.published.each(&:hide!)
      end

      transitions from: :published, to: :hidden
    end
  end

  belongs_to :prefecture
  has_many :casts, dependent: :destroy
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :favorite_people, dependent: :destroy
  has_many :staffs, as: :resource, dependent: :destroy
  has_many :users, through: :favorite_people

  def favorites
    favorite_people
  end

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
end
