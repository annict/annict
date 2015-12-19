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
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#  aasm_state       :string           default("published"), not null
#
# Indexes
#
#  index_people_on_name           (name) UNIQUE
#  index_people_on_prefecture_id  (prefecture_id)
#

class Person < ActiveRecord::Base
  extend Enumerize
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      after do
        cast_participations.each(&:hide!)
        staff_participations.each(&:hide!)
      end

      transitions from: :published, to: :hidden
    end
  end

  enumerize :gender, in: [:male, :female, :other], default: :other

  belongs_to :prefecture
  has_many :cast_participations, dependent: :destroy
  has_many :staff_participations, dependent: :destroy
end
