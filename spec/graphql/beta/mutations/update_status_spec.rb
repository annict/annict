# frozen_string_literal: true

describe Beta::Mutations::UpdateStatus do
  let(:user) { create :registered_user }
  let(:anime) { create :anime }
  let(:token) { create(:oauth_access_token) }
  let(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let(:anime_id) { Beta::AnnictSchema.id_from_object(anime, anime.class) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation($workId: ID!) {
          updateStatus(input: {
            workId: $workId,
            state: WANNA_WATCH
          }) {
            work {
              title
            }
          }
        }
      GRAPHQL
    end
    let(:variables) { {workId: anime_id} }

    it "ステータスが変更できること" do
      expect(Status.count).to eq 0

      result = Beta::AnnictSchema.execute(query, variables: variables, context: context)
      expect(result["errors"]).to be_nil

      expect(Status.count).to eq 1

      status = user.statuses.first
      expect(status.kind).to eq "wanna_watch"
      expect(status.work_id).to eq anime.id

      expect(result.dig("data", "updateStatus", "work", "title")).to eq anime.title
    end
  end
end
