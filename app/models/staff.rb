# == Schema Information
#
# Table name: staffs
#
#  id                 :integer          not null, primary key
#  email              :string           default(""), not null
#  encrypted_password :string           default(""), not null
#  sign_in_count      :integer          default(0), not null
#  current_sign_in_at :datetime
#  last_sign_in_at    :datetime
#  current_sign_in_ip :string
#  last_sign_in_ip    :string
#  created_at         :datetime
#  updated_at         :datetime
#
# Indexes
#
#  index_staffs_on_email  (email) UNIQUE
#

class Staff < ActiveRecord::Base
  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :omniauthable, :recoverable,
  # :rememberable, :registerable and :validatable
  devise :database_authenticatable, :trackable
end
