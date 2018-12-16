# frozen_string_literal: true

Mutations::CreateReview = GraphQL::Relay::Mutation.define do
  name "CreateReview"

  input_field :workId, !types.ID
  input_field :title, types.String
  input_field :body, !types.String
  WorkRecord::STATES.each do |state|
    input_field state.to_s.camelcase(:lower).to_sym, EnumTypes::RatingState
  end
  input_field :shareTwitter, types.Boolean
  input_field :shareFacebook, types.Boolean

  return_field :review, ObjectTypes::Review

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    work = Work.published.find_by_graphql_id(inputs[:workId])

    review = work.work_records.new do |r|
      r.user = ctx[:viewer]
      r.work = work
      r.title = inputs[:title]
      r.body = inputs[:body]
      WorkRecord::STATES.each do |state|
        r.send("#{state}=".to_sym, inputs[state.to_s.camelcase(:lower).to_sym]&.downcase)
      end
      r.oauth_application = ctx[:doorkeeper_token].application
    end
    ctx[:viewer].setting.attributes = {
      share_review_to_twitter: inputs[:shareTwitter] == true,
      share_review_to_facebook: inputs[:shareFacebook] == true
    }

    service = NewWorkRecordService.new(ctx[:viewer], review, ctx[:viewer].setting)
    service.via = "graphql_api"
    service.app = ctx[:doorkeeper_token].application
    service.ga_client = ctx[:ga_client]
    service.logentries = ctx[:logentries]

    service.save!

    {
      review: service.work_record
    }
  }
end
