# frozen_string_literal: true

class UserSlotsQuery
  class OrderProperty
    def initialize(field_ = nil, direction_ = nil)
      @field_ = field_
      @direction_ = direction_
    end

    def field
      field_&.to_s&.downcase&.to_sym.presence || :created_at
    end

    def direction
      direction_&.to_s&.downcase&.to_sym.presence || :asc
    end

    private

    attr_reader :field_, :direction_
  end

  # @param user [User]
  # @param slots [Slot::ActiveRecord_Relation]
  # @param watched [Boolean, nil]
  # @param order [OrderProperty]
  #
  # @return [Slot::ActiveRecord_Relation]
  def initialize(
    user,
    slots,
    watched: nil,
    order: OrderProperty.new
  )
    @user = user
    @slots = slots
    @watched = watched
    @order = order
  end

  def call
    collection = user_slots.preload(:channel, work: :work_image, episode: :work)
    order_collection(collection)
  end

  private

  attr_reader :user, :watched, :order

  def user_slots
    slots = @slots
    slots = slots.where(program_id: library_entries.pluck(:program_id))
    slots = slots.where.not(episode_id: library_entries.pluck(:watched_episode_ids).flatten) if watched == false
    slots
  end

  def library_entries
    @library_entries ||= user.library_entries.where.not(program_id: nil)
  end

  def order_collection(collection)
    return collection.order(:created_at) unless order

    collection.order(order.field => order.direction)
  end
end
