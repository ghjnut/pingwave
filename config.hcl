target_group "cloud" {
  prefix = "cloud"
  interval = 5

  target "ec2.us-east-1" {
    address = "ec2.us-east-1.amazonaws.com"
  }
  target "ec2.us-east-2" {
    address = "ec2.us-east-2.amazonaws.com"
  }
  target "ec2.us-west-1" {
    address = "ec2.us-west-1.amazonaws.com"
  }
  target "ec2.us-west-2" {
    address = "ec2.us-west-2.amazonaws.com"
  }
  target "gce.us-central1" {
    address = "us-central1-gce.cloudharmony.net"
  }
  target "gce.us-east1" {
    address = "us-east1-gce.cloudharmony.net"
  }
  target "gce.us-east4" {
    address = "us-east4-gce.cloudharmony.net"
  }
  target "gce.us-west1" {
    address = "us-west1-gce.cloudharmony.net"
  }
  target "gce.us-west2-a" {
    address = "us-west2-a-gce.cloudharmony.net"
  }
  #target "test.bad" {
  #  address = "193.132.32.5"
  #}
}

# Declare a target group with a name
target_group "websites" {
  interval = 3
  # a custom ping interval for this group
  # A prefix for the statsd metric for this group
  prefix = "websites"
  # A name for the target. This becomes the statsd metric
  target "youtube" {
    address = "youtube.com"
  }
  target "jupiter_broadcasting" {
    address = "jupiterbroadcasting.com"
  }
  target "twitch" {
    address = "twitch.tv"
  }
}
