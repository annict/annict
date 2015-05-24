class WorkDecorator < Draper::Decorator
  include Diffable

  delegate_all

  def to_diffable_resource
    hash = {}

    white_list = %w(
      season_id sc_tid title media official_site_url wikipedia_url
      twitter_username twitter_hashtag released_at released_at_about
    )

    white_list.each do |column_name|
      hash[column_name] = get_diffable_work(column_name, object.send(column_name))
    end

    hash
  end
end
