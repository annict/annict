# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for dynamic methods in `Canary::Mutations::UpdateEpisodeRecord`.
# Please instead update this file by running `bin/tapioca dsl Canary::Mutations::UpdateEpisodeRecord`.


class Canary::Mutations::UpdateEpisodeRecord
  sig do
    params(
      record_id: ::String,
      rating: T.nilable(::String),
      comment: T.nilable(::String),
      share_to_twitter: T.nilable(T::Boolean)
    ).returns(T.untyped)
  end
  def resolve(record_id:, rating: T.unsafe(nil), comment: T.unsafe(nil), share_to_twitter: T.unsafe(nil)); end
end
