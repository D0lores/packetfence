package pf::services::manager::haproxy_db;
=head1 NAME

pf::services::manager::haproxy_db add documentation

=cut

=head1 DESCRIPTION

pf::services::manager::haproxy_db

=cut

use strict;
use warnings;
use Moo;

use List::MoreUtils qw(uniq);

use pf::log;
use pf::util;
use pf::cluster;
use pf::config qw(
    %Config
    $OS
    @listen_ints
    @dhcplistener_ints
    $management_network
    @portal_ints
);
use pf::file_paths qw(
    $generated_conf_dir
    $install_dir
    $conf_dir
    $var_dir
    $captiveportal_templates_path
);

use pf::constants qw($TRUE $FALSE);

extends 'pf::services::manager::haproxy';

has '+name' => (default => sub { 'haproxy-db' } );

has '+haproxy_config_template' => (default => sub { "$conf_dir/haproxy-db.conf" });

our $host_id = $pf::config::cluster::host_id;
tie our %clusters_hostname_map, 'pfconfig::cached_hash', 'resource::clusters_hostname_map';

sub generateConfig {
    my ($self,$quick) = @_;
    my $logger = get_logger();
    my ($package, $filename, $line) = caller();

    my %tags;
    $tags{'template'} = $self->haproxy_config_template;
    $tags{'http'} = '';
    $tags{'mysql_backend'} = '';
    $tags{'var_dir'} = $var_dir;
    $tags{'conf_dir'} = $var_dir.'/conf';
    if ($OS eq 'debian') {
        $tags{'os_path'} = '/etc/haproxy/errors/';
    } else {
         $tags{'os_path'} = '/usr/share/haproxy/';
    }
    
    $tags{'management_ip'}
        = defined( $management_network->tag('vip') )
        ? $management_network->tag('vip')
        : $management_network->tag('ip');

    my $i = 0;
    my @mysql_backend;

    if ($cluster_enabled) {
        my $management_ip = pf::cluster::management_cluster_ip();
        if ($self->isSlaveMode()) {
            if ($self->getDBMaster()) {
                 push(@mysql_backend, $self->getDBMaster());
            }
            push(@mysql_backend, $tags{'management_ip'});
        } else {
            @mysql_backend = map { $_->{management_ip} } pf::cluster::mysql_servers();
        }
        $tags{'management_ip_frontend'} = <<"EOT";
frontend  management_ip
    bind $management_ip:3306
    mode tcp
    option tcplog
    default_backend             mysql
EOT
    } else {
        @mysql_backend = split(',', $Config{database_advanced}{other_members});
        push(@mysql_backend, $tags{'management_ip'});
        $tags{'management_ip_frontend'} = '';
    }
    foreach my $mysql_back (@mysql_backend) {
        # the second server (the one without the VIP) will be the prefered MySQL server
        if ($i == 0) {
        $tags{'mysql_backend'} .= <<"EOT";
server MySQL$i $mysql_back:3306 check
EOT
        } else {
        $tags{'mysql_backend'} .= <<"EOT";
server MySQL$i $mysql_back:3306 check backup
EOT
        }
    $i++;
    }
    
    $tags{captiveportal_templates_path} = $captiveportal_templates_path;
    parse_template( \%tags, $self->haproxy_config_template, "$generated_conf_dir/".$self->name.".conf" );

    return 1;
}

sub isManaged {
    my ($self) = @_;
    my $name = $self->name;
    if (isenabled($pf::config::Config{'services'}{$name})) {
        if ($cluster_enabled && $self->isSlaveMode()) {
            return $TRUE;
        }
        return $cluster_enabled;
    } else {
        return 0;
    }
}

sub isSlaveMode {
    my ($self) = @_;
    if (defined(${pf::config::cluster::getClusterConfig($clusters_hostname_map{$host_id})}{CLUSTER}{masterslavemode}) && ${pf::config::cluster::getClusterConfig($clusters_hostname_map{$host_id})}{CLUSTER}{masterslavemode} eq 'SLAVE' ) {
        return $TRUE;
    }
}

sub getDBMaster {
    if (defined(${pf::config::cluster::getClusterConfig(${pf::config::cluster::getClusterConfig($clusters_hostname_map{$host_id})}{CLUSTER}{masterdb})}{CLUSTER}{management_ip})) {
        return ${pf::config::cluster::getClusterConfig(${pf::config::cluster::getClusterConfig($clusters_hostname_map{$host_id})}{CLUSTER}{masterdb})}{CLUSTER}{management_ip};
    } else {
        return $FALSE;
    }
}

=head1 AUTHOR

Inverse inc. <info@inverse.ca>



=head1 COPYRIGHT

Copyright (C) 2005-2020 Inverse inc.

=head1 LICENSE

This program is free software; you can redistribute it and::or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301,
USA.

=cut

1;
