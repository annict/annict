# == Schema Information
#
# Table name: people
#
#  id               :integer          not null, primary key
#  prefecture_id    :integer
#  name             :string           not null
#  name_kana        :string
#  nickname         :string
#  gender           :string
#  url              :string
#  wikipedia_url    :string
#  twitter_username :string
#  birthday         :date
#  blood_type       :string
#  height           :integer
#  aasm_state       :string           default("published"), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
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
  include PersonCommon

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      after do
        casts.each(&:hide!)
        staffs.each(&:hide!)
      end

      transitions from: :published, to: :hidden
    end
  end

  belongs_to :prefecture
  has_many :casts, dependent: :destroy
  has_many :draft_casts, dependent: :destroy
  has_many :draft_staffs, dependent: :destroy
  has_many :staffs, dependent: :destroy
end
