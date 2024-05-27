# typed: false
# frozen_string_literal: true

describe "POST /api/internal/unstars" do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  context "when user unstars a character" do
    let!(:character_favorite) { create(:character_favorite, user: user) }
    let!(:data) { {starrable_type: "Character", starrable_id: character_favorite.character_id} }

    it "responses 200" do
      post api("/api/internal/unstars", data)
      expect(response.status).to eq(201)
    end

    it "removes a record" do
      expect { post api("/api/internal/unstars", data) }
        .to change { user.favorite_characters.count }.from(1).to(0)
    end
  end

  context "when user unstars a organization" do
    let!(:organization_favorite) { create(:organization_favorite, user: user) }
    let!(:data) { {starrable_type: "Organization", starrable_id: organization_favorite.organization_id} }

    it "responses 200" do
      post api("/api/internal/unstars", data)
      expect(response.status).to eq(201)
    end

    it "removes a record" do
      expect { post api("/api/internal/unstars", data) }
        .to change { user.favorite_organizations.count }.from(1).to(0)
    end
  end

  context "when user unstars person" do
    let!(:person_favorite) { create(:person_favorite, user: user) }
    let!(:data) { {starrable_type: "Person", starrable_id: person_favorite.person_id} }

    it "responses 200" do
      post api("/api/internal/unstars", data)
      expect(response.status).to eq(201)
    end

    it "removes a record" do
      expect { post api("/api/internal/unstars", data) }
        .to change { user.favorite_people.count }.from(1).to(0)
    end
  end
end
