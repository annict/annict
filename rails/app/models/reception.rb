# typed: false

class Reception < ApplicationRecord
  belongs_to :channel
  belongs_to :user

  def self.initial?(reception)
    count == 1 && first.id == reception.id
  end
end
