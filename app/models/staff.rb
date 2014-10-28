# == Schema Information
#
# Table name: staffs
#
#  id                 :integer          not null, primary key
#  email              :string(255)      default(""), not null
#  encrypted_password :string(255)      default(""), not null
#  sign_in_count      :integer          default(0), not null
#  current_sign_in_at :datetime
#  last_sign_in_at    :datetime
#  current_sign_in_ip :string(255)
#  last_sign_in_ip    :string(255)
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
