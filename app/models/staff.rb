# == Schema Information
#
# Table name: staffs
#
#  id                 :integer          not null, primary key
#  email              :string(510)      default(""), not null
#  encrypted_password :string(510)      default(""), not null
#  sign_in_count      :integer          default("0"), not null
#  current_sign_in_at :datetime
#  last_sign_in_at    :datetime
#  current_sign_in_ip :string(510)
#  last_sign_in_ip    :string(510)
#  created_at         :datetime
#  updated_at         :datetime
#
# Indexes
#
#  staffs_email_key  (email) UNIQUE
#

class Staff < ActiveRecord::Base
  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :omniauthable, :recoverable,
  # :rememberable, :registerable and :validatable
  devise :database_authenticatable, :trackable
end
