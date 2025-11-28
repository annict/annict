# typed: false
# frozen_string_literal: true

class ChannelGroup < ApplicationRecord
  include Unpublishable

  has_many :channels
end
