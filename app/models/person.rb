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
#
# Indexes
#
#  index_people_on_name           (name) UNIQUE
#  index_people_on_prefecture_id  (prefecture_id)
#

class Person < ActiveRecord::Base
  belongs_to :prefecture
  has_many :org_participations, dependent: :destroy
  has_many :organizations, through: :org_participations
  has_many :work_participations, dependent: :destroy
  has_many :works, through: :work_participations
end
