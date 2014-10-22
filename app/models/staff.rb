class Staff < ActiveRecord::Base
  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :omniauthable, :recoverable,
  # :rememberable, :registerable and :validatable
  devise :database_authenticatable, :trackable
end
