# frozen_string_literal: true

describe Connections::ActivityConnection do
  before do
    @user = create(:user)
    @episode_record = create(:episode_record, user: @user)
    @status = create(:status, user: @user)
    create(:activity, user: @user, recipient: @episode_record.episode, trackable: @episode_record)

    @id = GraphQL::Schema::UniqueWithinType.encode(@user.class.name, @user.id)
    query_string = <<~GRAPHQL
      query {
        node(id: "#{@id}") {
          id
          ... on User {
            username
            activities {
              edges {
                node {
                  ... on Record {
                    comment
                  }
                  ... on Status {
                    state
                  }
                }
              }
            }
          }
        }
      }
    GRAPHQL
    @res = AnnictSchema.execute(query_string)
  end

  it "fetches activities" do
    expected = {
      data: {
        node: {
          id: @id,
          username: @user.username,
          activities: {
            edges: [
              {
                node: {
                  state: @status.kind.upcase.to_s
                }
              },
              {
                node: {
                  comment: @episode_record.comment
                }
              }
            ]
          }
        }
      }
    }
    expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
  end
end
