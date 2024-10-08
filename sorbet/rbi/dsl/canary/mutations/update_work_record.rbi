# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for dynamic methods in `Canary::Mutations::UpdateWorkRecord`.
# Please instead update this file by running `bin/tapioca dsl Canary::Mutations::UpdateWorkRecord`.


class Canary::Mutations::UpdateWorkRecord
  sig do
    params(
      record_id: ::String,
      rating_overall: T.nilable(::String),
      rating_animation: T.nilable(::String),
      rating_music: T.nilable(::String),
      rating_story: T.nilable(::String),
      rating_character: T.nilable(::String),
      comment: T.nilable(::String),
      share_to_twitter: T.nilable(T::Boolean)
    ).returns(T.untyped)
  end
  def resolve(record_id:, rating_overall: T.unsafe(nil), rating_animation: T.unsafe(nil), rating_music: T.unsafe(nil), rating_story: T.unsafe(nil), rating_character: T.unsafe(nil), comment: T.unsafe(nil), share_to_twitter: T.unsafe(nil)); end
end
