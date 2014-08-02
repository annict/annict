class Activity < ActiveRecord::Base
  belongs_to :recipient, polymorphic: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user
end