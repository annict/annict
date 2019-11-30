# frozen_string_literal: true

class UserSlotsQuery
  # @param user [User]
  # @param slots [Slot::ActiveRecord_Relation]
  # @param status_kinds [Array<Symbol>]
  # @param watched [Boolean, nil]
  # @param order [OrderProperty]
  #
  # @return [Slot::ActiveRecord_Relation]
  def initialize(
    user,
    slots,
    status_kinds: LatestStatus.kind.values.map(&:to_sym),
    watched: nil,
    order: OrderProperty.new
  )
    @user = user
    @slots = slots
    @status_kinds = status_kinds
    @watched = watched
    @order = order
  end

  def call
    collection = user_slots
    order_collection(collection)
  end

  private

  attr_reader :user, :slots, :status_kinds, :watched, :order

  def user_slots
    return slots.none if user_episodes.blank?

    id_pair = channel_works.pluck(:channel_id, :work_id).map { |ary| "(#{ary[0]}, #{ary[1]})" }.join(",")
    id_pair_sql = <<-SQL
      SELECT id FROM slots WHERE
        (channel_id, work_id) IN (VALUES #{id_pair})
    SQL

    # 過去の放送を取得しないようにするため、直近の放送分のみ取得する
    slot_sql = <<-SQL
      WITH ranked_slots AS (
        SELECT id, episode_id,
          dense_rank() OVER (
            PARTITION BY episode_id ORDER BY started_at DESC
          ) AS episode_rank
        FROM slots
        WHERE
          id IN (#{id_pair_sql})
      )
      SELECT id FROM ranked_slots
      WHERE
        episode_id IN (#{user_episodes.pluck(:id).join(',')}) AND
        episode_rank = 1;
    SQL
    slots_ids = slots.find_by_sql(slot_sql)

    slots.where(id: slots_ids.pluck(:id))
  end

  def user_episodes
    @user_episodes ||= UserEpisodesQuery.new(
      user,
      Episode.without_deleted.where(work_id: latest_statuses.pluck(:work_id)),
      watched: watched
    ).call
  end

  def channel_works
    @channel_works ||= user.channel_works.where(work: latest_statuses.pluck(:work_id))
  end

  def latest_statuses
    @latest_statuses ||= user.latest_statuses.with_kind(*status_kinds)
  end

  def order_collection(collection)
    return collection.order(:created_at) unless order

    collection.order(order.field => order.direction)
  end
end
